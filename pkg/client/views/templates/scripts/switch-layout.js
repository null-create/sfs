function switchLayout(view) {
  const gridView = document.getElementById('file-grid');
  const tableView = document.getElementById('file-table');
  if (view === 'grid') {
      gridView.style.display = 'flex';
      tableView.style.display = 'none';
  } else if (view === 'table') {
      gridView.style.display = 'none';
      tableView.style.display = 'block';
  }
}