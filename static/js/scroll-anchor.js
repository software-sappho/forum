document.addEventListener("DOMContentLoaded", function () {
  const hash = window.location.hash;
  if (hash) {
    const target = document.getElementById(hash.substring(1));
    if (target) {
      target.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  }
});
