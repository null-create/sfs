document
.getElementById("upload-form")
.addEventListener("submit", function (event) {
  event.preventDefault(); // Prevent the default form submission

  const fileInput = document.getElementById("folder-input");
  const formData = new FormData();

  if (fileInput.files.length > 0) {
    // Get the folder path from the form
    const file = fileInput.files[0];
    formData.append("folder-path", file);

    // Send the form data (folder path) to the server
    fetch(`/upload/${file}`, {
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