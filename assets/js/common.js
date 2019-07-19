const ENTER_KEY = 13;
const ESCAPE_KEY = 27;

class ListingPage {
    constructor() {
        this.addButton = document.getElementById('add-button');
        this.itemList = document.getElementById('item-list');
        this.emptyList = document.getElementById('empty-list');
        this.addDialog = this.createAddDialog();

        this.addButton.addEventListener('click', () => this.addDialog.open());
        this.addDialog.onAdd = (item) => {
            this.addItem(item).catch((err) => handleError(err));
        }
    }

    updateData(items) {
        if (items.length === 0) {
            this.itemList.style.display = 'none';
            this.emptyList.style.display = 'block';
            return;
        }
        this.itemList.innerHTML = '';
        items.forEach((item) => {
            this.addListItem(item);
        });
        this.itemList.style.display = 'block';
        this.emptyList.style.display = 'none';
    }

    addListItem(item) {
        const element = this.createListItem(item);
        element.addEventListener('click', () => {
            this.selectedListItem(item);
        });
        const deleteButton = document.createElement('button');
        deleteButton.className = 'delete-button';
        deleteButton.textContent = 'Delete';
        deleteButton.addEventListener('click', (e) => {
            e.stopPropagation();
            this.deleteItem(item).catch((err) => handleError(err));
        });
        element.appendChild(deleteButton);
        this.itemList.appendChild(element);
    }

    createAddDialog() {
        throw new Error('override this in a subclass');
    }

    createListItem(item) {
        throw new Error('override this in a subclass');
    }

    selectedListItem(item) {
        throw new Error('override this in a subclass');
    }

    addItem(item) {
        throw new Error('override this in a subclass');
    }

    deleteItem(item) {
        throw new Error('override this in a subclass');
    }
}

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
                this.close();
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
