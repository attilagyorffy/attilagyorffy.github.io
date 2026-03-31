const faviconActive = "/images/favicon-active.svg";
const faviconInactive = "/images/favicon-inactive.svg";
const favicon = document.getElementById("favicon");

document.addEventListener("visibilitychange", () => {
    if (document.hidden) {
        favicon.href = faviconInactive;
    } else {
        favicon.href = faviconActive;
    }
});
