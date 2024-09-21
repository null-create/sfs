const buttonClasses = {
  ".add-button": 'add-button', // key= button class, value=dropdown ID
  ".remove-button": 'remove-button', 
  ".type-button": 'type-button', 
  ".modified-button": 'modified-button',
  ".location-button": 'location-button',
};

document.addEventListener("DOMContentLoaded", () => {
  // gather a set of attributes about each of the buttons on the page

  // set event listeners for all buttons, including clearing them if the
  // user clicks elsewhere on the page that isn't part of the button or dropdown menu
  for (const [className, id] of Object.entries(buttonClasses)) {
    console.log(`${className}: ${id}`);
    const button = document.getElementsByClassName(className)
    if (!button) {
      console.error(`couldn't find button using class name: ${className}`)
    } else {
      button.addEventListener("click", () => {
        document.getElementById(id).classList.toggle("show");
      });
    }
  }
});

// clear all dropdown items if the user clicks somewhere on the page
window.onclick = (event) => {
  for (const [className, id] of Object.entries(buttonClasses)) {
    console.log(`${className}: ${id}`);
    if (!event.target.matches(className)) {
      // TODO: investigate whether dropdown-content should be a universal class name
      var dropdowns = document.getElementsByClassName("dropdown-content"); 
      for (var i = 0; i < dropdowns.length; i++) {
        var openDropdown = dropdowns[i];
        if (openDropdown.classList.contains('show')) {
          openDropdown.classList.remove('show');
        }
      }
    }
  }
};