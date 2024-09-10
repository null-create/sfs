// Check if the server is up
function checkServerStatus(serverURL) {
  console.log("serverURL: " + serverURL);
  const statusText = document.getElementById('status-text'); 
  fetch(serverURL)
  .then(response => {
    if (response.ok) {
      statusText.textContent = 'online';
      statusText.classList.remove('offline');
      statusText.classList.add('online');
    } else {
      throw new Error('Server offline');
    }
  })
  .catch(error => {
    statusText.textContent = 'offline';
    statusText.classList.remove('online');
    statusText.classList.add('offline');
  });
}