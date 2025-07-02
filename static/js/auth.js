document.addEventListener('DOMContentLoaded', function() {
    // Handle registration form
    const registerForm = document.querySelector('form[action="/register"]');
    if (registerForm) {
        registerForm.addEventListener('submit', function(e) {
            const password = document.getElementById('password').value;
            const confirmPassword = document.getElementById('confirm_password').value;
            
            if (password !== confirmPassword) {
                e.preventDefault();
                alert('Passwords do not match!');
            }
        });
    }

    // Handle login form
    const loginForm = document.querySelector('form[action="/login"]');
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;
            
            if (!email || !password) {
                e.preventDefault();
                alert('Please fill in all fields!');
            }
        });
    }

    // Handle password requirements visibility
    const passwordField = document.getElementById('password');
    const confirmPasswordField = document.getElementById('confirm_password');
    const passwordRequirements = document.getElementById('passwordRequirements');
    
    if (passwordField && passwordRequirements) {
        // Show requirements when password field is focused
        passwordField.addEventListener('focus', function() {
            passwordRequirements.classList.remove('hidden');
        });
        
        // Hide requirements when password field loses focus
        passwordField.addEventListener('blur', function() {
            passwordRequirements.classList.add('hidden');
        });
    }
    
    if (confirmPasswordField && passwordRequirements) {
        // Also show requirements when confirm password field is focused
        confirmPasswordField.addEventListener('focus', function() {
            passwordRequirements.classList.remove('hidden');
        });
        
        // Hide requirements when confirm password field loses focus
        confirmPasswordField.addEventListener('blur', function() {
            passwordRequirements.classList.add('hidden');
        });
    }
}); 