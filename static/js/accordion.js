function toggleAccordion(contentId) {
    const content = document.getElementById(contentId);
    const arrow = document.getElementById(contentId.replace('content', 'arrow'));
    
    if (content.classList.contains('hidden')) {
        content.classList.remove('hidden');
        arrow.classList.add('rotate-180');
    } else {
        content.classList.add('hidden');
        arrow.classList.remove('rotate-180');
    }
} 