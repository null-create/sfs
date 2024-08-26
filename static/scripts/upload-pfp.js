function uploadPfp() {
  const profilePicInput = document.getElementById("profile-pic-upload");
  if (profilePicInput) {
    profilePicInput.addEventListener("change", function () {
      const file = this.files[0];
      if (file) {
        const formData = new FormData();
        formData.append("profilePic", file);
        // Send to the client server
        fetch("/user/upload-pfp", {
          method: "POST",
          body: formData,
        })
          .then((response) => {
            if (response.ok) {
              console.log(response.json());
              return;
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
      else {
        console.error("no file data received:");
        return;
      }
    });
  } else {
    console.error("profile-pic-upload element not found.");
  }
}

document.addEventListener("DOMContentLoaded", uploadPfp);
