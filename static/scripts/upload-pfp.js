const fileInput = document.getElementById("profile-pic-upload");
fileInput.addEventListener("change", (event) => {
  const form = document.getElementById("upload-form");
  const formData = new FormData(form);
  const searchParams = new URLSearchParams(formData);
  const fetchOptions = {
    method: form.method,
  };

  if (form.method.toLowerCase() === 'post') {
    if (form.enctype === 'multipart/form-data') {
      fetchOptions.body = formData;
    } else {
      fetchOptions.body = searchParams;
    }
  } else {
    url.search = searchParams;
  }

  console.log("fetching...");
  fetch("/user/upload-pfp", fetchOptions)
  .then((response) => {
    if (response.ok) {
      console.log("picture updated successfully")
      window.location.href = "/user"
    }
  })
  .catch((error) => {
    console.error("picture update failed: ", error)
  });

  event.preventDefault();
});