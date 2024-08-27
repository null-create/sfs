function addItems() {
  document.getElementById("submit-folder-path").addEventListener("click", () => {
    const spinner = document.getElementById("spinner");
    spinner.style.display = "block";
    
    const msgElement = document.getElementById("response");
    msgElement.style.display = "none";

    const pathInput = document.getElementById("folder-path-input").value;
    if (!pathInput) {
      alert("please select a path to a file or folder")
    }    
    console.log("path input: " + pathInput);
    fetch("/add/new", {
      method: "POST",
      body: pathInput
    })
    .then((response) => {
      spinner.style.display = "none";
      if (!response.ok) {
        console.error("response status: " + response.status);
      }
      msgElement.style.display = "block";
      msgElement.textContent = "Item(s) added successfully";
      msgElement.classList.add("success"); 
    })
    .catch((error) => {
      msgElement.style.display = "block";
      msgElement.textContent = error.message;
      msgElement.classList.add("error"); 
      console.error("error:", error);
    });
  });
}

document.addEventListener("DOMContentLoaded", addItems);