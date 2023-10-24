package client

// ---- file monitoring operations

// populates c.Monitor.Events with listener goroutines for all
// files in the client's drive.
func (c *Client) WatchFiles() error {
	return c.Monitor.WatchAll(c.Drive.Root.Path)
}

// add a file listener to the map if the file isn't already present.
// will be a no-op if its already being watched.
func (c *Client) WatchFile(filePath string) {
	if _, exists := c.Monitor.Events[filePath]; !exists {
		c.Monitor.WatchFile(filePath)
	}
}

// stop all event listeners for this client
func (c *Client) StopMonitoring() error { return nil }

// main event loop that coordinates sync operations after
// receiving a file event from the listener.

// TODO: this should be part of a larger data structure
// that is coordinating (or at least keeping track of) all
// the watcher/listener event goroutines
func (c *Client) EventHandler() error { return nil }
