(function () {

    class RoutePage {
        constructor() {
            this.svgElement = document.getElementsByTagName('svg')[0];
            this.labels = this.svgElement.getElementsByTagNameNS(
                'http://www.w3.org/2000/svg',
                'text',
            );

            this.listItems = LIST_DATA.map((x) => createListItem(x, 'div'));
            this.currentListItem = document.getElementById('current-list-item');
            this.currentIndex = 0;
            this.showCurrentListItem();

            this.prevButton = document.getElementById('prev-button');
            this.prevButton.addEventListener('click', () => {
                if (this.currentIndex > 0) {
                    this.currentIndex--;
                    this.showCurrentListItem();
                }
            });
            this.nextButton = document.getElementById('next-button');
            this.nextButton.addEventListener('click', () => {
                if (this.currentIndex + 1 < this.listItems.length) {
                    this.currentIndex++;
                    this.showCurrentListItem();
                }
            });
        }

        showCurrentListItem() {
            const next = this.currentListItem.nextElementSibling;
            this.currentListItem.parentNode.removeChild(this.currentListItem);
            this.currentListItem = this.listItems[this.currentIndex];
            next.parentElement.insertBefore(this.currentListItem, next);

            this.emphasizeLabel(LIST_DATA[this.currentIndex].zone);
        }

        emphasizeLabel(text) {
            for (let i = 0; i < this.labels.length; ++i) {
                const label = this.labels[i];
                if (label.textContent.trim() === text) {
                    label.style.fontWeight = 'bolder';
                    label.style.fill = 'red';
                } else {
                    label.style.fontWeight = 'normal';
                    label.style.fill = 'black';
                }
            }
        }
    }

    window.addEventListener('load', () => {
        new RoutePage();
    });

})();