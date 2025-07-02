// Custom dropdown functionality
document.addEventListener('DOMContentLoaded', function() {
    const dropdownButton = document.getElementById('dropdownDefaultButton');
    const dropdown = document.getElementById('dropdown');
    const selectedCategorySpan = document.getElementById('selectedCategory');
    const categoryInput = document.getElementById('category');
    const categoryOptions = document.querySelectorAll('.category-option');

    // Toggle dropdown
    dropdownButton.addEventListener('click', function() {
        dropdown.classList.toggle('hidden');
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', function(event) {
        if (!dropdownButton.contains(event.target) && !dropdown.contains(event.target)) {
            dropdown.classList.add('hidden');
        }
    });

    // Handle category selection
    categoryOptions.forEach(option => {
        option.addEventListener('click', function() {
            const value = this.getAttribute('data-value');
            selectedCategorySpan.textContent = value;
            categoryInput.value = value;
            dropdown.classList.add('hidden');
        });
    });
}); 