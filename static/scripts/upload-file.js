function uploadFile(newFilesEndpoint) {
  console.log("new files endpoint: " + newFilesEndpoint);
  
  const msgElement = document.getElementById("response");
  msgElement.style.display = "none";

  const fileInput = document.getElementById("upload-form");
  fileInput.addEventListener("change",  () => {
    const file = this.files[0];
    if (file) {
      console.log("uploading file: " + file);
      const formData = new FormData();
      formData.append("newFile", file);
      const destFolderName = getDestFolderName();
      if (!destFolderName) {
        throw new Error("no destination folder specified");
      }
      formData.append("destDir", destFolderName);
      
      // upload the file
      fetch(newFilesEndpoint, {
        method: "POST",
        body: formData,
        headers: {
          folder: destFolderName,
        },
      })
      .then((response) => {
        if (response.ok) {
          msgElement.style.display = "block";
          msgElement.textContent = `${file} uploaded successfully`;
          msgElement.classList.add("success"); 
          console.log(response.json());
          return;
        }
      })
      .catch((error) => {
        msgElement.style.display = "block";
        msgElement.textContent = error.message;
        msgElement.classList.add("error"); 
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
