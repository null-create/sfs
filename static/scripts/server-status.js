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
        offlineStatus();
      }
    })
    .catch((error) => {
      console.log(error)
      offlineStatus();
    });
};
