// DOM Elements
const loadingState = document.getElementById('loadingState');
const errorState = document.getElementById('errorState');
const contentState = document.getElementById('contentState');
const apodDate = document.getElementById('apodDate');
const apodTitle = document.getElementById('apodTitle');
const apodImage = document.getElementById('apodImage');
const apodVideo = document.getElementById('apodVideo');
const apodExplanation = document.getElementById('apodExplanation');
const apodCopyright = document.getElementById('apodCopyright');

// UI State
const showLoading = () => {
  loadingState.classList.remove('hidden');
  errorState.classList.add('hidden');
  contentState.classList.add('hidden');
};

const showError = () => {
  loadingState.classList.add('hidden');
  errorState.classList.remove('hidden');
  contentState.classList.add('hidden');
};

const showContent = () => {
  loadingState.classList.add('hidden');
  errorState.classList.add('hidden');
  contentState.classList.remove('hidden');
};

// Format date for display
const formatDate = (dateStr) =>
  new Date(dateStr).toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });

// Render APOD data to the page
const displayAPOD = (data) => {
  console.log('APOD data received:', data);

  apodTitle.textContent = data.title;
  apodDate.textContent = formatDate(data.date);
  apodExplanation.textContent = data.explanation;
  apodCopyright.textContent = data.copyright
    ? `© ${data.copyright}`
    : '';

  apodVideo.innerHTML = '';
  if (data.media_type === 'image') {
    apodImage.src = data.url;
    apodImage.alt = data.title;
    apodImage.classList.remove('hidden');
    apodVideo.classList.add('hidden');
  } else if (data.media_type === 'video') {
    const videoId = new URL(data.url).searchParams.get('v');
    apodVideo.innerHTML = `
      <iframe
        class="w-full h-full"
        src="https://www.youtube.com/embed/${videoId}"
        style="border: none;"
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
        allowfullscreen
      ></iframe>
    `;
    apodImage.classList.add('hidden');
    apodVideo.classList.remove('hidden');
  } else {
    console.warn('Unsupported media type:', data.media_type);
  }

  showContent();
};

// Fetch APOD from backend proxy
const fetchAPOD = async (date) => {
  showLoading();

  const params = new URLSearchParams();
  if (date) params.append('date', date);

  try {
    const response = await fetch(`/api/nasa-apod?${params.toString()}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data = await response.json();
    displayAPOD(data);
  } catch (err) {
    console.error('Failed to load APOD:', err);
    showError();
  }
};

// Load today's APOD on page load
document.addEventListener('DOMContentLoaded', () => {
  fetchAPOD();
});
