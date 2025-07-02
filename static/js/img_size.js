document.addEventListener("DOMContentLoaded", () => {
    const form = document.getElementById("create-post-form");
    const imageInput = document.getElementById("image");
    const imageError = document.getElementById("image-error");
    const maxFileSize = 20 * 1024 * 1024; // 20 MB

    form.addEventListener("submit", function (event) {
        // If no file is selected, do not perform any validation and let the form submit.
        // The back-end will handle the case of a missing image.
        if (imageInput.files.length === 0) {
            imageError.classList.add("hidden");
            imageError.textContent = "";
            return; 
        }

        const file = imageInput.files[0];
        if (file) {
            if (file.size > maxFileSize) {
                event.preventDefault(); // Stop form submission
                imageError.textContent = "Image is too large (max 20 MB). Please choose a smaller file.";
                imageError.classList.remove("hidden");
            } else {
                imageError.classList.add("hidden");
                imageError.textContent = "";
            }
        }
    });
}); 