// Profile picture upload functionality
document.addEventListener('DOMContentLoaded', function() {
  console.log('Profile JS loaded');
  
  const profilePicInput = document.getElementById('profile_picture');
  const fileChosen = document.getElementById('file-chosen');
  const profileForm = document.getElementById('profile-form');
  const maxFileSize = 500 * 1024; // 500 KB

  console.log('Elements found:', {
    profilePicInput: !!profilePicInput,
    fileChosen: !!fileChosen,
    profileForm: !!profileForm
  });

  if (!profilePicInput || !fileChosen || !profileForm) {
    console.error('Required elements not found');
    return;
  }

  profilePicInput.addEventListener('change', function () {
    console.log('File input changed');
    const file = this.files?.[0];

    if (!file) {
      console.log('No file selected');
      fileChosen.textContent = 'Click the edit icon to change picture';
      return;
    }

    console.log('File selected:', file.name, file.size, file.type);
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif'];

    // Update UI to show file name
    fileChosen.textContent = `Selected: ${file.name}`;
    fileChosen.classList.remove('text-red-600');

    if (!allowedTypes.includes(file.type)) {
      console.log('Invalid file type');
      fileChosen.textContent = 'Invalid file type. Use JPG, PNG, or GIF.';
      fileChosen.classList.add('text-red-600');
      return;
    }

    if (file.size > maxFileSize) {
      console.log('File too large');
      fileChosen.textContent = `Image file size is too big! Maximum size is 500 KB. Your file is ${(file.size / 1024).toFixed(1)} KB.`;
      fileChosen.classList.add('text-red-600');
      return;
    }

    // If file is valid, auto-submit the form
    console.log('File is valid, submitting form');
    fileChosen.textContent = 'Uploading...';
    profileForm.submit();
  });
}); 