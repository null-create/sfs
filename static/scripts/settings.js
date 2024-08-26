function submitSettings() {
  document.getElementById("loading-spinner").style.display = "block"; // Show loading spinner
  const settings = {
    theme: document.getElementById("theme").value,
    notifications: document.getElementById("notifications").checked,
    serverSync: document.getElementById("server-sync").checked,
  };

  fetch("/settings", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(settings),
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error("Error updating settings");
      }
    })
    .then((data) => {
      console.log("Settings updated successfully");
      document.getElementById("loading-spinner").style.display = "none";
      alert("Settings updated successfully")
    })
    .catch((error) => {
      console.error("Error:", error);
      document.getElementById("loading-spinner").style.display = "none";
      alert(error)
    });
}
