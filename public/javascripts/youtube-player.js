(function () {
    var trigger = document.querySelector(".yt-trigger");
    var lightbox = document.getElementById("lightbox");
    if (!trigger || !lightbox) return;

    var player = lightbox.querySelector(".lightbox-player");
    var closeBtn = lightbox.querySelector(".lightbox-close");
    var videoId = trigger.dataset.videoId;
    var videoTitle = trigger.dataset.videoTitle || "YouTube video";

    function open() {
        var iframe = document.createElement("iframe");
        iframe.src =
            "https://www.youtube-nocookie.com/embed/" +
            videoId +
            "?autoplay=1&rel=0";
        iframe.setAttribute(
            "allow",
            "accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture",
        );
        iframe.setAttribute("allowfullscreen", "");
        iframe.setAttribute("title", videoTitle);
        player.appendChild(iframe);
        lightbox.setAttribute("aria-hidden", "false");
        document.body.style.overflow = "hidden";
    }

    function close() {
        lightbox.setAttribute("aria-hidden", "true");
        player.innerHTML = "";
        document.body.style.overflow = "";
    }

    trigger.addEventListener("click", function (e) {
        e.preventDefault();
        open();
    });

    closeBtn.addEventListener("click", close);

    lightbox.addEventListener("click", function (e) {
        if (e.target === lightbox) close();
    });

    document.addEventListener("keydown", function (e) {
        if (lightbox.getAttribute("aria-hidden") !== "false") return;
        if (e.key === "Escape") close();
    });
})();
