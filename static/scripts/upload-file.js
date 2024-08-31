document.addEventListener("DOMContentLoaded", () => {
  const uploadButton = document.getElementById("upload-button");
  const fileUploadInput = document.getElementById("file-upload");
  const destinationFolderSelect = document.getElementById("destination-folder");
  const responseDiv = document.getElementById("response");

  uploadButton.addEventListener("click", (event) => {
    event.preventDefault();

    const selectedFolder = destinationFolderSelect.value;
    const file = fileUploadInput.files[0];
    console.log("selectedFolder: " + selectedFolder);
    console.log("file: " + file);

    if (!file) {
      alert("Please select a file to upload.");
      return;
    }
    if (!selectedFolder) {
      alert("Please select a destination folder.");
      return;
    }

    const formData = new FormData();
    formData.append("file", file);
    formData.append("destFolder", selectedFolder);

    fetch("/upload", {
      method: "POST",
      body: formData,
    })
    .then((response) => {
      if (response.ok) {
        responseDiv.textContent = "File(s) uploaded successfully.";
      } else {
        responseDiv.textContent = "Error uploading file: " + JSON.stringify(response);
      }
    })
    .catch((error) => {
      responseDiv.textContent = `Error: ${error.message}`;
      console.error("Upload error:", error);
    });
  });

  // Handling drag and drop events on the dropzone
  const dropzone = document.getElementById("upload-form");

  dropzone.addEventListener("dragover", (event) => {
    event.preventDefault(); 
  });

  dropzone.addEventListener("drop", (event) => {
    event.preventDefault();
    const files = event.dataTransfer.files;
    if (files.length > 0) {
      fileUploadInput.files = files; 
      alert("File ready to upload.");
    }
  });
});
