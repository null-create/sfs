function submitSettings() {
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
