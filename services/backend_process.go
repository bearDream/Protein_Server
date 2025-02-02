package services

import "time"

// A service that always runs on the back end
func BackendProcess() {
	a_processor := NewAlphaProcessor(3)
	i_processor := NewItasserProcessor(3)

	go FetchPDB()

	go func() {
		fetchTicker := time.NewTicker(24 * time.Hour)
		defer fetchTicker.Stop()
		for range fetchTicker.C {
			go FetchPDB()
		}
	}()

	// Start processing
	for {
		// Check queue every minute
		a_processor.Process()
		i_processor.Process()
		time.Sleep(time.Minute)
	}
}
