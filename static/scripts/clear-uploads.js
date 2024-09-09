function removeAllUploads() {
  const dropzoneInstance = Dropzone.forElement("#upload-form");
  dropzoneInstance.removeAllFiles(true);
}