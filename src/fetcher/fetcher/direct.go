package fetcher

import (
	"instagram"
	"time"
	"utils/log"
)

// direct activity: accept PM invites; read && parse them
func directActivity(api *instagram.Instagram, rndTimeout time.Duration) {

	for {
		// get pending shiet
		// and accept it
		err := acceptDirectThreads(api)
		if err != nil {
			log.Error(err)
			continue
		}

		time.Sleep(rndTimeout)
		return
	}
}

func acceptDirectThreads(api *instagram.Instagram) error {
	pendingInbox, err := api.PendingInbox()
	if err != nil {
		return err
	}

	for _, thread := range pendingInbox.Inbox.Threads {
		if thread.Pending {
			log.Debug("Trying to approve")
			resp, err := api.DirectThreadAction(thread.ThreadID, instagram.ActionApprove)
			if err != nil {
				return err
			}

			_ = resp
			// @TODO check if result is really says request was accepted
		}
	}

	return nil
}

func checkNewMessages(api *instagram.Instagram) error {

	// get non-pending shiet
	// check which threads got updated since last time
	// get them
	chats, err := api.RankedRecipients()
	if err != nil {
		return err
	}

	_ = chats
	return nil
}
