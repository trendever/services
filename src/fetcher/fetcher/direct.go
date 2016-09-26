package fetcher

import (
	"instagram"
	"time"
	"utils/log"
)

// direct activity: accept PM invites; read && parse them
func directActivity(api *instagram.Instagram, rndTimeout time.Duration) {

	for {

		err := checkNewMessages(api)
		if err != nil {
			log.Error(err)
			continue
		}

		time.Sleep(rndTimeout)
	}
}

func checkNewMessages(api *instagram.Instagram) error {

	// get non-pending shiet
	// check which threads got updated since last time
	// get them
	resp, err := api.Inbox("")
	if err != nil {
		return err
	}

	if resp.PendingRequestsTotal > 0 {
		_, err := api.DirectThreadApproveAll()
		// do nothing else now
		return err
	}

	for _, thread := range resp.Inbox.Threads {

		err := processThread(api, thread.ThreadID)
		if err != nil {
			return err
		}
	}

	return nil
}

func processThread(api *instagram.Instagram, threadID string) error {

	resp, err := api.DirectThread(threadID, "")
	if err != nil {
		return err
	}

	_ = resp

	return nil
}
