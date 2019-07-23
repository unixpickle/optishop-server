const ENTER_KEY = 13;
const ESCAPE_KEY = 27;
const MIN_OVERLAY_LOADER_TIME = 300;

class ListingPage {
    constructor() {
        this.addButton = document.getElementById('add-button');
        this.itemList = document.getElementById('item-list');
        this.emptyList = document.getElementById('empty-list');
        this.addDialog = this.createAddDialog();

        this.addButton.addEventListener('click', () => this.addDialog.open());
        this.addDialog.onAdd = (item) => this.addItem(item);
    }

    updateData(items) {
        if (items.length === 0) {
            // Empty the list so that if we add a new item
            // with addListItem, old items are not there.
            this.itemList.innerHTML = '';
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
            const hideLoader = showOverlayLoader();
            this.deleteItem(item).catch(handleError).finally(hideLoader);
        });
        element.appendChild(deleteButton);
        this.itemList.appendChild(element);

        // Incase the list used to be empty.
        this.itemList.style.display = 'block';
        this.emptyList.style.display = 'none';
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
        this.content = this.element.getElementsByClassName('dialog-content')[0];
        this.searchBox = document.getElementById('search-box');
        this.searchButton = document.getElementById('search-button');
        this.searchResults = document.getElementById('search-results');
        this.closeAddButton = document.getElementById('close-add-button');
        this.loader = createBasicLoader();

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
                this.searchBox.blur();
                this.search();
            }
        })

        // Callback which is called when an item is selected.
        this.onAdd = (item) => null;
    }

    open() {
        document.body.classList.add('doing-search');
        this.instanceNum++;
        this.searchBox.value = '';
        this.searchResults.innerHTML = '';
        this.hideLoader();
        this.element.style.display = 'block';
        this.searchBox.focus();
    }

    close() {
        document.body.classList.remove('doing-search');
        this.instanceNum++;
        this.element.style.display = 'none';
    }

    search() {
        this.instanceNum++;
        const query = this.searchBox.value;
        const instanceNum = this.instanceNum;

        this.searchResults.innerHTML = '';
        this.showLoader();
        this.fetchSearchResults(query).then((results) => {
            if (instanceNum !== this.instanceNum) {
                return;
            }
            this.hideLoader();
            results.forEach((item) => {
                const elem = this.createListItem(item);
                elem.addEventListener('click', () => {
                    const hideLoader = showOverlayLoader();
                    this.onAdd(item)
                        .then(() => this.close())
                        .catch(handleError)
                        .finally(hideLoader);
                });
                this.searchResults.appendChild(elem);
            });
        }).catch((err) => {
            this.hideLoader();
            handleError(err);
        });
    }

    hideLoader() {
        if (this.loader.parentElement) {
            this.loader.parentElement.removeChild(this.loader);
        }
    }

    showLoader() {
        if (!this.loader.parentElement) {
            this.content.appendChild(this.loader);
        }
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

function createBasicLoader() {
    const element = document.createElement('div');
    element.className = 'loader';
    return element;
}

function showOverlayLoader() {
    const background = document.createElement('div');
    background.className = 'loader-overlay';
    const container = document.createElement('div');
    container.className = 'loader-overlay-container';
    container.appendChild(createBasicLoader());
    background.appendChild(container);
    document.body.appendChild(background);
    const shownTime = new Date().getTime();
    return () => {
        const elapsed = Math.max(0, new Date().getTime() - shownTime);
        if (elapsed > MIN_OVERLAY_LOADER_TIME) {
            document.body.removeChild(background);
            return;
        }
        setTimeout(() => document.body.removeChild(background), MIN_OVERLAY_LOADER_TIME - elapsed);
    }
}
