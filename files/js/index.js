const today = new Date();
const slot = document.getElementById("today");
if (slot) {
  slot.innerText = today.toLocaleDateString();
}