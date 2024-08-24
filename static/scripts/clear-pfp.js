document.getElementById("clear-profile-pic-button").addEventListener("click", function() {
  // Send a request to clear the profile picture on the server
  fetch("/user/clear-pfp", {
    method: "POST",
  })
    .then((response) => {
      if (response.ok) {
        // Set the profile picture back to the default image
        document.getElementById("user-profile-pic").src = "/assets/default_profile_pic.jpg";
        console.log("Profile picture cleared successfully");
      } else {
        throw new Error("Failed to clear profile picture");
      }
    })
    .catch((error) => {
      console.error("Error:", error);
    });
});
