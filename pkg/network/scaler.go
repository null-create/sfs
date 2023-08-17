package network

/*
ScaleBatches() generates a network profile file,
then chooses a a batch size limit used for uploading or downloading.
*/
func ScaleBatches() {
	p := ProfileNetwork()

	// establish network upload/download batch sizes
	PickMAX(p)
}
