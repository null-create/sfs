function clearPfp() {
  document.getElementById("clear-profile-pic-button")
  .addEventListener("click", () => {
    fetch("/user/clear-pfp", {
      method: "POST",
    })
    .then((response) => {
      if (response.ok) {
        console.log("Profile picture cleared successfully");
        window.location.href = "/user"
      } else {
        throw new Error("Failed to clear profile picture");
      }
    })
    .catch((error) => {
      console.error("Error:", error);
      alert("Failed to clear profile picture: " + error.message);
    });
  });  
}

document.addEventListener("DOMContentLoaded", clearPfp);