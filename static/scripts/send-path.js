function sendPathForDiscovery() {
  document.getElementById("submit-folder-path").addEventListener("click", function () {
    const pathInput = document.getElementById("folder-path-input").value;
    if (!pathInput) {
      alert("please select a path to a file or folder")
    }
    console.log("path input: " + pathInput);
    fetch("/add/discover", {
      method: "POST",
      body: pathInput 
    })
    .then((response) => {
      if (response.ok) {
        alert("Folder added successfully");
      }
      else {
        console.error("response status: " + response.status)
      }
    })
    .catch((error) => {
      console.error("error:", error);
    });
  });
}

document.addEventListener("DOMContentLoaded", sendPathForDiscovery);