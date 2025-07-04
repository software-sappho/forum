 // Cursor
    let div = document.getElementById("move-div");
    let title = document.getElementById("title");
    //Detect touch device
    function isTouchDevice() {
        try {
            //We try to create Touch Event (it would fail for desktops and throw error)
            document.createEvent("TouchEvent");
            return true;
        } catch (e) {
            return false;
        }
    }
    const move = (e) => {
        //Try catch to avoid any errors for touch screens(Error thrown when user doesn't move his finger)
        try {
            /*
            PageX and PageY return the position of clients cursor from the top left of screen
            */
            var x = !isTouchDevice() ? e.pageX : e.touches[0].pageX;
            var y = !isTouchDevice() ? e.pageY : e.touches[0].pageY;
        } catch (error) {}
        //Set left and top of div based on mouse position
        div.style.left = x + "px";
        div.style.top = y + "px";
    };
    //For mouse
    document.addEventListener("mousemove", move);
    //For touch
    document.addEventListener("touchmove", move);