function removeFile(fileID) {
  fetch("/files/delete", {
    method: "DELETE",
    body: fileID
  })
  .then((response) => {
    if (response.ok) {
      window.location.href = "/"
    }
  })
  .catch((error) => {
    console.error("Error:", error);
    alert(error.message);
  });
}