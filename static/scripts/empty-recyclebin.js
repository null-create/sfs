function emptyBin() {
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