function openFileLoc(fileID) {
  fetch(`/files/i/${fileID}/open-loc`)
  .then((response) => {
    if (response.ok) {
      console.log("success")
    }
  })
  .catch((error) => {
    console.error("error:", error);
    alert(error.message);
  });
}