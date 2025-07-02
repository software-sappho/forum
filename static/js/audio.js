document.addEventListener('DOMContentLoaded', function() {
    const audio = document.getElementById('backgroundAudio');
    const audioControl = document.getElementById('audioControl');
    const audioIcon = document.getElementById('audioIcon');
    
    // Set initial volume
    audio.volume = 0.5;
    
    // Start muted by default
    audio.muted = true;
    audioIcon.src = "/static/img/muteicon.png";
    
    // Toggle mute/unmute on button click
    audioControl.addEventListener('click', function() {
        if (audio.muted) {
            // Try to play when unmuting
            const playPromise = audio.play();
            
            if (playPromise !== undefined) {
                playPromise.then(() => {
                    // Playback started successfully
                    audio.muted = false;
                    audioIcon.src = "/static/img/unmuteicon.png";
                }).catch(error => {
                    // Playback failed
                    console.log('Playback failed:', error);
                    // Keep muted state
                    audio.muted = true;
                    audioIcon.src = "/static/img/muteicon.png";
                });
            }
        } else {
            audio.muted = true;
            audioIcon.src = "/static/img/muteicon.png";
        }
    });
});