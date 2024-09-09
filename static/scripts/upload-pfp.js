const fileInput = document.getElementById("profile-pic-upload");
fileInput.addEventListener("change", (event) => {
  event.preventDefault();

  const form = document.getElementById("upload-form");
  const formData = new FormData(form);

  fetch("/user/upload-pfp", {
    method: "POST",
    body: formData
  })
  .then((response) => {
    if (response.ok) {
      console.log("picture updated successfully")
      window.location.href = "/user"
    } else {
      console.error("error uploading picture: " + response);
      window.location.href = "/user"
    }
  })
  .catch((error) => {
    alert("error uploading picture: " + error)
    console.error("picture update failed: ", error)
  });
});