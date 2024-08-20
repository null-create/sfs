function sendPathForDiscovery() {
  document
    .getElementById("upload-form")
    .addEventListener("submit", function (event) {
      event.preventDefault(); // Prevent the default form submission

      const pathInput = document.getElementById("folder-input");
      const formData = new FormData();

      if (pathInput.files.length > 0) {
        const folderPath = pathInput.files[0]; // Get the folder path from the form
        formData.append("folder-path", folderPath);

        // Send the form data (folder path) to the client server endpoint
        fetch(`/add/discover/${folderPath}`, {
          method: "POST",
          body: formData,
        })
          .then((response) => response.json())
          .then((data) => {
            console.log("Success:", data);
            alert("Folder added successfully");
          })
          .catch((error) => {
            console.error("Error:", error);
            alert("Failed add the folder");
          });
      } else {
        alert("Please select a folder to add");
      }
    });
}
document.addEventListener("DOMContentLoaded", sendPathForDiscovery);