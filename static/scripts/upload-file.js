function uploadFile(newFilesEndpoint) {
  console.log("new files endpoint: " + newFilesEndpoint);
  
  const fileInput = document.getElementById("upload-form");
  fileInput.addEventListener("change",  () => {
    const file = this.files[0];
    if (file) {
      const formData = new FormData();
      formData.append("newFile", file);
      const destFolderName = getDestFolderName();
      if (!destFolderName) {
        throw new Error("no destination folder specified");
      }
      formData.append("destDir", destFolderName);

      fetch(newFilesEndpoint, {
        method: "POST",
        body: formData,
        headers: {
          folder: destFolderName,
        },
      })
        .then((response) => {
          if (response.ok) {
            console.log(response.json());
            return;
          } else {
            throw new Error("Failed to upload profile picture");
          }
        })
        .catch((error) => {
          console.error("Error:", error);
        });
    } else {
      console.error("no file data received");
    }
  });
}

function getDestFolderName() {
  const selectedFolder = document.getElementById("folder-selection").value;
  console.log("Selected folder: ", selectedFolder);
  return selectedFolder;
}

document.addEventListener("DOMContentLoaded", uploadFile);