const ENTER_KEY = 13;
const ESCAPE_KEY = 27;

class AddDialog {
    constructor() {
        this.element = document.getElementById('add-dialog');
        this.searchBox = document.getElementById('search-box');
        this.searchButton = document.getElementById('search-button');
        this.searchResults = document.getElementById('search-results');
        this.closeAddButton = document.getElementById('close-add-button');

        // Used to prevent old requests from affecting the
        // add dialog if it is closed and re-opened.
        this.instanceNum = 0;

        this.closeAddButton.addEventListener('click', () => this.close());
        window.addEventListener('keyup', (e) => {
            if (e.which === ESCAPE_KEY) {
                this.addDialog.close();
            }
        });

        this.searchButton.addEventListener('click', () => this.search());
        this.searchBox.addEventListener('keyup', (e) => {
            if (e.which === ENTER_KEY) {
                this.search();
            }
        })

        // Callback which is called when an item is selected.
        this.onAdd = (item) => null;
    }

    open() {
        this.instanceNum++;
        this.searchBox.value = '';
        this.searchResults.innerHTML = '';
        this.element.style.display = 'block';
    }

    close() {
        this.instanceNum++;
        this.element.style.display = 'none';
    }

    search() {
        this.instanceNum++;
        const query = this.searchBox.value;
        const instanceNum = this.instanceNum;

        this.searchResults.innerHTML = '';
        this.fetchSearchResults(query).then((results) => {
            if (instanceNum !== this.instanceNum) {
                return;
            }
            results.forEach((item) => {
                const elem = this.createListItem(item);
                elem.addEventListener('click', () => {
                    this.close();
                    this.onAdd(item);
                });
                this.searchResults.appendChild(elem);
            });
        }).catch((err) => {
            handleError(err);
        });
    }

    fetchSearchResults(query) {
        throw new Error('override this in a subclass');
    }

    createListItem(item) {
        throw new Error('override this in a subclass');
    }
}

function handleError(err) {
    alert('Error: ' + err);
}
