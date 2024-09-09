function submitSearch(event) {
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
document.getElementById("search-form").addEventListener("submit", submitSearch);
