class PageManager {
    constructor() {
        this.pages = Array.prototype.slice.apply(document.getElementsByClassName('page'));
    }

    displayPage(name) {
        this.pages.forEach((page) => {
            if (page.id == name + '-page') {
                page.style.display = 'block';
            } else {
                page.style.display = 'none';
            }
        });
    }
}
