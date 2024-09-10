function openFileLoc(fileID) {
  const url = `/files/i/${fileID}/open-loc`
  console.log("fetching: "+ url)
  fetch(url)
  .then((response) => {
    if (response.ok) {
      window.location.href = `/files/i/${fileID}`
    }
  })
  .catch((error) => {
    console.error("Error:", error);
    alert(error.message);
  });
}