/*
Web client UI functionality
*/ 

// ------- misc utilities --------------------------------


// redirect to another page
const redirectToPage = (url) => {
  window.location.href = url;
}

// hide upper search bar when on home page
const hideSearchBar = () => {
  const searchBar = document.getElementById("search");
  if (searchBar) {
    searchBar.style.display = "none";
  }
}

// TODO: switch themes (dark or light)

// ------- buttons ---------------------------------------

// add button
document.addEventListener("DOMContentLoaded", () => {
  const addBtn = document.getElementById("dropdown-btn");
  if (addBtn) {
    addBtn.addEventListener("click", () => {
      document.getElementById("dropdown-menu").classList.toggle("show");
    });
  }
});  

// close the add dropdown if the user clicks outside of it
window.onclick = (event) => {
  if (!event.target.matches('.add-button')) {
    let dropdowns = document.getElementsByClassName("dropdown-content");
    for (let i = 0; i < dropdowns.length; i++) {
      let openDropdown = dropdowns[i];
      if (openDropdown.classList.contains('show')) {
        openDropdown.classList.remove('show');
      }
    }
  }
}

const closeMenu = (event, buttonCn, menuCn) => {
  if (event.target.matches(buttonCn)) {
    let dropdowns = document.getElementsByClassName(menuCn);
    dropdowns.forEach((e) => {
      if (e.classList.contains('show')) {
        e.classList.remove('show');
      }
    })
  }
}


// ------- uploads ---------------------------------------

const addItems = () => {
  const folderPath = document.getElementById("submit-folder-path")
  if (folderPath) {
    folderPath.addEventListener("click", () => {
      const spinner = document.getElementById("spinner");
      spinner.style.display = "block";
      
      const msgElement = document.getElementById("response");
      msgElement.style.display = "none";
  
      const pathInput = document.getElementById("folder-path-input").value;
      if (!pathInput) {
        alert("please select a path to a file or folder")
        return;
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
        } else{
          msgElement.style.display = "block";
          msgElement.textContent = "Item(s) added successfully";
          msgElement.classList.add("success"); 
        }
      })
      .catch((error) => {
        msgElement.style.display = "block";
        msgElement.textContent = error.message;
        msgElement.classList.add("error"); 
        console.error("error:", error);
      });
    });
  }

}

document.addEventListener("DOMContentLoaded", addItems);

const removeAllUploads = () => {
  const dropzoneInstance = Dropzone.forElement("#dropzone-form");
  dropzoneInstance.removeAllFiles(true);
}

// dropzone upload handling
document.addEventListener("DOMContentLoaded", () => {
  const uploadButton = document.getElementById("upload-button");
  if (uploadButton) {
    const fileUploadInput = document.getElementById("file-upload");
    const destinationFolderSelect = document.getElementById("destination-folder");
    const responseDiv = document.getElementById("response");
  
    uploadButton.addEventListener("click", (event) => {
      event.preventDefault();
  
      const selectedFolder = destinationFolderSelect.value;
      const file = fileUploadInput.files[0];
      console.log("selectedFolder: " + selectedFolder);
      console.log("file: " + file);
  
      if (!file) {
        alert("Please select a file to upload.");
        return;
      }
      if (!selectedFolder) {
        alert("Please select a destination folder.");
        return;
      }
  
      const formData = new FormData();
      formData.append("file", file);
      formData.append("destFolder", selectedFolder);
  
      fetch("/upload", {
        method: "POST",
        body: formData,
      })
      .then((response) => {
        if (response.ok) {
          responseDiv.textContent = "File(s) uploaded successfully.";
        } else {
          responseDiv.textContent = "Error uploading file: " + JSON.stringify(response);
        }
      })
      .catch((error) => {
        responseDiv.textContent = `Error: ${error.message}`;
        console.error("Upload error:", error);
      });
    });
  
    // Handling drag and drop events on the dropzone
    const dropzone = document.getElementById("dropzone-form");
    if (dropzone) {
      dropzone.addEventListener("dragover", (event) => {
        event.preventDefault(); 
      });
      dropzone.addEventListener("drop", (event) => {
        event.preventDefault();
        const files = event.dataTransfer.files;
        if (files.length > 0) {
          fileUploadInput.files = files; 
          alert("File ready to upload.");
        }
      });
    }
  }
});


// -------- check remote server status --------------------

const checkServerStatus = (serverURL) => {
  const statusText = document.getElementById("status-text");
  if (serverURL === "") {
    console.log("serverURL is empty");
    statusText.textContent = "offline";
    statusText.classList.remove("online");
    statusText.classList.add("offline");
    return
  }
  console.log("serverURL: " + serverURL);

  const onlineStatus = () => {
    statusText.textContent = "online";
    statusText.classList.remove("offline");
    statusText.classList.add("online");
  };

  const offlineStatus = () => {
    statusText.textContent = "offline";
    statusText.classList.remove("online");
    statusText.classList.add("offline");
  };

  fetch(serverURL)
  .then((response) => {
    if (response.ok) {
      onlineStatus();
    } else {
      console.log("server response: ", JSON.stringify(response));
      offlineStatus();
    }
  })
  .catch((error) => {
    console.log(error)
    offlineStatus();
  });
};

// --------- user page ---------------------------------------

// upload PFP
document.addEventListener("DOMContentLoaded", () => {
  const fileInput = document.getElementById("profile-pic-upload");
  if (fileInput) {
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
        } else {
          console.error("error uploading picture: " + response);
        }
        redirectToPage("/user");
      })
      .catch((error) => {
        alert("error uploading picture: " + error)
        console.error("picture update failed: ", error)
      });
    });
  }
});


const clearPfp = () => {
  fetch("/user/clear-pfp", {
    method: "POST",
  })
  .then((response) => {
    if (response.ok) {
      console.log("Profile picture cleared successfully");
      redirectToPage("/user");
    } else {
      throw new Error("Failed to clear profile picture");
    }
  })
  .catch((error) => {
    console.error("Error:", error);
    alert("Failed to clear profile picture: " + error.message);
  }); 
}

const pfpButton = document.getElementById("clear-profile-pic-button")
if (pfpButton) {
  pfpButton.addEventListener("click", clearPfp);
}


// --------- recycle bin page --------------------------------

const emptyBin = () => {
  if (confirm("WARNING: This will permanently delete *all* items in the SFS recycle bin. Proceed?")){
    fetch("/empty", {
      method: "DELETE"
    })
    .then((response) => {
      if (response.ok) {
        window.location.href = "/recycled"
      }
    })
    .catch((error) => {
      console.error("Error:", error);
      alert(error.message);
    });
  } else {
    window.location.href = "/"
  }
}

// -------- search page ---------------------------------------

const submitSearch = (event) => {
  event.preventDefault();  // Prevent form from reloading the page
  const searchItem = document.getElementById('search-input').value;
  console.log("search query: " + searchItem);

  fetch("/search", {
    method: 'POST',
    body: searchItem,
  })
  .then((response) => {
    if (!response.ok) {
      console.error('response was not ok: ', response.status);
    } else {
      console.log('Search request sent successfully.');
      window.location.href = `/search?searchQuery=${encodeURIComponent(searchItem)}`;
    }
  })
  .catch((error) => {
    console.error('Error during search:', error);
    alert(error);
  });
}

// Add the event listener to the form itself
document.addEventListener('DOMContentLoaded', () => {
  const searchForm = document.getElementById("search-form");
  if (searchForm) {
    searchForm.addEventListener("submit", submitSearch);
  }
});

// --------- edit page ----------------------------------------

const submitEdits = () => {
  const editForm = document.getElementById("edit-info-form");
  if (editForm) {
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
        body: formData,
      })
      .then((response) => {
        if (response.ok) {
          console.log("Success")
          redirectToPage("/user");
        }
      })
      .catch((error) => {
        alert("Error: " + error);
      });
    });
  }
}

document.addEventListener("DOMContentLoaded", submitEdits)

// --------- settings page ------------------------------------

const submitSettings = () => {
  // const theme = document.getElementById("theme").value
  const serverSync = document.getElementById("server-sync").checked;
  const backupDir = document.getElementById("local-backup-dir").value;
  const clientPort = document.getElementById("client-port").value;
  const syncDelay = document.getElementById("sync-delay").value;

  // TODO: handle theme on the client side

  const settings = {
    CLIENT_LOCAL_BACKUP: serverSync,
    CLIENT_BACKUP_DIR: backupDir,
    CLIENT_PORT: clientPort,
    EVENT_BUFFER_SIZE: syncDelay
  };

  console.log("sending settings to server: ", JSON.stringify(settings));

  fetch("/settings", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(settings),
  })
  .then((response) => {
    if (!response.ok) {
      throw new Error(response.status + ": " + response.statusText);
    }
  })
  .then((data) => {
    console.log("Settings updated successfully");
    alert("Settings updated successfully")
  })
  .catch((error) => {
    console.error("Error:", error);
    alert(error)
  });
}


// -------- file page -----------------------------------------

const openFileLoc = (fileID) => {
  fetch(`/files/i/${fileID}/open-loc`)
  .then((response) => {
    if (response.ok) {
      console.log("success")
    }
  })
  .catch((error) => {
    console.error("error:", error);
    alert(error.message);
  });
}

const removeFile = (fileID) => {
  fetch("/files/delete", {
    method: "DELETE",
    body: fileID
  })
  .then((response) => {
    if (response.ok) {
      window.location.href = "/"
    }
  })
  .catch((error) => {
    console.error("Error:", error);
    alert(error.message);
  });
}