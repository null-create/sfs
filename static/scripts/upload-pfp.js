function uploadPfp() {
  document
    .getElementById("profile-pic-upload")
    .addEventListener("change", function () {
      const file = this.files[0];
      if (file) {
        const formData = new FormData();
        formData.append("profilePic", file); // Append the file to FormData
        fetch("/upload-profile-pic", {
          method: "POST",
          body: formData,
        })
          .then((response) => {
            if (response.ok) {
              return response.json();
            } else {
              throw new Error("Failed to upload profile picture");
            }
          })
          .then((data) => {
            // Assuming the server responds with the image URL
            document.getElementById("user-profile-pic").src = data.imageUrl;
            console.log("Profile picture uploaded successfully");
          })
          .catch((error) => {
            console.error("Error:", error);
          });
      }
    });
}