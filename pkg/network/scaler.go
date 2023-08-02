package network

/*
Scalar takes a network profile file, then chooses a range of possible file batch sizes used for uploading or downloading.

Upload/Download pairs will be mapped against a set of file batch sizes.

Batch sizes should be limited to file SIZE rather than file quantity, so as not to accidentally create
a huge batch of huge files (and clog the user's network).

*/
