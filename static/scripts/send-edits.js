function submitEdits() {
  const editForm = document.getElementById("edit-info-form");
  editForm.addEventListener("submit", () => {
    const newName = document.getElementById("name-edit").value;
    const newUsername = document.getElementById("username-edit").value;
    const newEmail = document.getElementById("email-edit").value;

    const formData = new FormData();
    formData.append("name", newName);
    formData.append("username", newUsername);
    formData.append("email", newEmail);

    fetch("/user/edit", {
      method: "POST",
      body: formData
    })
    .then((response) => {
      if (response.ok) {
        alert("Information has been updated successfully");
        window.location.href = "/user"
      }
    })
    .catch((error) => {
      alert("Error: " + error);
    });
  });

}

document.addEventListener("DOMContentLoaded", submitEdits)