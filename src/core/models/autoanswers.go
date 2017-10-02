package models

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kljensen/snowball/english"
	"github.com/kljensen/snowball/russian"
	//"github.com/kljensen/snowball/french"
	//"github.com/kljensen/snowball/spanish"
	"common/db"
	"common/log"
	"github.com/qor/validations"
	"proto/chat"
	"strings"
	"sync"
	"unicode"
)

var AnswersSupportedLanguages []string

var stemmers = map[string]func(string, bool) string{
	"russian": russian.Stem,
	"english": english.Stem,
	//"french":  french.Stem,
	//"spanish": spanish.Stem,
}

func init() {
	for lang := range stemmers {
		AnswersSupportedLanguages = append(AnswersSupportedLanguages, lang)
	}
}

type AutoAnswer struct {
	db.Model
	Name string
	// one of SupportedLanguages
	Language string
	// comma-separated phases
	Dictionary string `gorm:"text"`
	// splited and steammed Dictionary
	preparedDictionary []string
	prepared           bool
	// chat template
	Template string `gorm:"text"`
}

var (
	// language -> []AutoAnswer
	autos map[string][]*AutoAnswer
	lock  sync.RWMutex
)

// loads autoanswers form db
func ReloadAnswers() error {
	var results []*AutoAnswer
	err := db.New().Find(&results).Error
	if err != nil {
		return fmt.Errorf("failed to load answers: %v", err)
	}
	lock.Lock()
	autos = map[string][]*AutoAnswer{}
	for _, auto := range results {
		if err = auto.Prepare(); err != nil {
			log.Errorf("failed to prepare dictionary '%v': %v", auto.Name, err)
		}
		autos[auto.Language] = append(autos[auto.Language], auto)
	}
	lock.Unlock()
	return nil
}

// splits text on non-letters and non-numbers runes, lowercases and stemmes words,
// joins all back with a single space
func PrepareText(text, language string) (string, error) {
	stemmer, ok := stemmers[language]
	if !ok {
		return "", fmt.Errorf("unsupported language '%v'", language)
	}
	words := strings.FieldsFunc(text, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})
	for i := range words {
		words[i] = stemmer(strings.ToLower(words[i]), false)
	}
	return strings.Join(words, " "), nil
}

// prepares dictionary
func (auto *AutoAnswer) Prepare() error {
	if auto.prepared {
		return nil
	}
	for _, phase := range strings.Split(auto.Dictionary, ",") {
		prepared, err := PrepareText(phase, auto.Language)
		if err != nil {
			return fmt.Errorf("failed to prepare phase '%v': %v", phase, err)
		}
		if len(prepared) == 0 {
			continue
		}
		auto.preparedDictionary = append(auto.preparedDictionary, prepared)
	}
	auto.prepared = true
	return nil
}

// returns true if any of phases in dictionary is presented in preparedText,
// Prepare() method should be called before this or result will be false,
// argument should be prepared with PrepareText as well
func (auto *AutoAnswer) Match(preparedText string) bool {
	for _, phase := range auto.preparedDictionary {
		if strings.Contains(preparedText, phase) {
			log.Debug("phase '%v' matchs text '%v'", phase, preparedText)
			return true
		}
	}
	return false
}

// reload Answers after any changes
func (auto *AutoAnswer) AfterCommit() {
	go func() {
		log.Error(ReloadAnswers())
	}()
}

func (auto *AutoAnswer) Validate(db *gorm.DB) {
	if auto.Name == "" {
		db.AddError(validations.NewError(auto, "Name", "Name should not be empty"))
	}

	ok := false
	for _, lang := range AnswersSupportedLanguages {
		if auto.Language == lang {
			ok = true
			break
		}
	}
	if !ok {
		db.AddError(validations.NewError(auto, "Language", "Unsupported language"))
	}

	err := auto.Prepare()
	if err != nil {
		db.AddError(validations.NewError(
			auto, "Dictionary",
			fmt.Sprintf("Failed to prepare dictionaty: %v", err),
		))
	}
}

func GenerateAnswers(text, language string, templatesContext interface{}) ([]string, error) {
	lock.RLock()
	suitable := autos[language]
	lock.RUnlock()

	var ret = []string{}
	if len(suitable) == 0 {
		return ret, nil
	}

	prepared, err := PrepareText(text, language)
	if err != nil {
		return nil, err
	}

	for _, auto := range suitable {
		if !auto.Match(prepared) {
			continue
		}
		answer, err := applyTemplate(auto.Template, templatesContext, false)
		if err != nil {
			log.Errorf("failed to execute Math() of AutoAnsver '%v' on text '%v'", auto.Name, text)
			continue
		}
		if len(answer) != 0 {
			ret = append(ret, answer)
		}
	}
	return ret, nil
}

func IsMessageAuto(msg *chat.Message) (bool, error) {
	for _, part := range msg.Parts {
		auto, err := IsMessagePartAuto(part)
		if err != nil {
			return false, err
		} else if auto {
			return true, nil
		}
	}

	return false, nil
}

func IsMessagePartAuto(part *chat.MessagePart) (bool, error) {

	if part.MimeType != "text/x-attrs" {
		return false, nil
	}

	var attrs map[string]interface{}

	err := json.Unmarshal([]byte(part.Content), &attrs)
	if err != nil {
		return false, err
	}

	if val, ok := attrs["type"]; ok {
		t, _ := val.(string)
		return t == "auto_answer", nil
	}

	return false, nil
}

func SendAutoAnswers(msg *chat.Message, lead *Lead) {
	var messages []*chat.Message
	for _, part := range msg.Parts {
		switch part.MimeType {
		case "text/plain":

		case "text/x-attrs":
			autoMsg, err := IsMessagePartAuto(part)
			if err != nil {
				log.Errorf("Warning, bad x-attrs: %v", err)
			} else if autoMsg {
				return
			}

		default:
			continue
		}
		// @TODO check message language somehow? possible set it on lead lvl
		// Генерируем автоответы по славорям
		answers, err := GenerateAnswers(part.Content, "russian", map[string]interface{}{
			"user": lead.Customer,
			"lead": lead,
		})
		if err != nil {
			log.Errorf("failed to generate autoanswers: %v", err)
		}

		//Составляем массив сообщений
		for _, answer := range answers {
			messages = append(messages, &chat.Message{
				UserId: uint64(SystemUser.ID),
				Parts: []*chat.MessagePart{
					{
						MimeType: "text/plain",
						Content:  answer,
					},
					{
						MimeType: "text/x-attrs",
						Content:  `{"type": "auto_answer"}`,
					},
				},
			})
		}
	}

	// Отправляем сообщения в чат
	err := SendChatMessages(lead.ConversationID, messages...)
	if err != nil {
		log.Errorf("failed to send messages to chat: %v", err)
	}
}
