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

func acceptDirectThreads(api *instagram.Instagram) error {
	pendingInbox, err := api.PendingInbox()
	if err != nil {
		return err
	}

	for _, thread := range pendingInbox.Inbox.Threads {
		if thread.Pending {

			// @TODO: various checks (for example, must be one-to-one chat)

			log.Debug("Approving message thread with ID=%v", thread.ThreadID)
			_, err := api.DirectThreadAction(thread.ThreadID, instagram.ActionApprove)
			// response status is checked inside instagram API and corresponding err is generated in this case
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func checkNewMessages(api *instagram.Instagram) error {

	// get non-pending shiet
	// check which threads got updated since last time
	// get them
	resp, err := api.Inbox()
	if err != nil {
		return err
	}

	if resp.PendingRequestsTotal > 0 {
		return acceptDirectThreads(api)
	}

	return nil
}
