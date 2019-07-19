(function () {

    const ENTER_KEY = 13;
    const ESCAPE_KEY = 27;

    class StoresPage {
        constructor() {
            this.addButton = document.getElementById('add-button');
            this.storesList = document.getElementById('stores-list');
            this.noStores = document.getElementById('no-stores');
            this.addDialog = new AddDialog();

            this.addButton.addEventListener('click', () => this.addDialog.open());
            this.addDialog.onAdd = (store) => this.addStore(store);

            this.data = null;
            this.updateData(window.STORES_DATA);

            window.addEventListener('keyup', (e) => {
                if (e.which === ESCAPE_KEY) {
                    this.addDialog.close();
                }
            });
        }

        updateData(data) {
            this.data = data;

            if (data.length === 0) {
                this.storesList.style.display = 'none';
                this.noStores.style.display = 'block';
                return;
            }
            this.storesList.innerHTML = '';
            data.forEach((datum) => {
                this.addListItem(datum);
            });
            this.storesList.style.display = 'block';
            this.noStores.style.display = 'none';
        }

        addStore(store) {
            const formData = 'source=' + encodeURIComponent(store.source) +
                '&signature=' + encodeURIComponent(store.signature) +
                '&data=' + encodeURIComponent(JSON.stringify(store.data));
            fetch('/api/addstore', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {
                    'content-type': 'application/x-www-form-urlencoded',
                },
                body: formData,
            }).then((x) => x.json()).then((data) => {
                if (data.error) {
                    handleError(data.error);
                    return;
                }
                store.id = data;
                this.addListItem(store);
            }).catch((err) => handleError(err));
        }

        deleteStore(id) {
            fetch('/api/removestore?store=' + encodeURIComponent(id), {
                credentials: 'same-origin',
            }).then((x) => x.json()).then((data) => {
                if (data.error) {
                    handleError(data.error);
                    return;
                }
                this.updateData(data);
            }).catch((err) => handleError(err));
        }

        addListItem(store) {
            const element = createListItem(store);
            element.addEventListener('click', () => {
                window.open('/list?store=' + encodeURIComponent(store.id));
            });
            const deleteButton = document.createElement('button');
            deleteButton.className = 'delete-button';
            deleteButton.textContent = 'Delete';
            deleteButton.addEventListener('click', (e) => {
                e.stopPropagation();
                this.deleteStore(store.id);
            });
            element.appendChild(deleteButton);
            this.storesList.appendChild(element);
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
            this.searchButton.addEventListener('click', () => this.search());
            this.searchBox.addEventListener('keyup', (e) => {
                if (e.which === ENTER_KEY) {
                    this.search();
                }
            })

            this.onAdd = () => null;
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
            fetch('/api/storequery?query=' + encodeURIComponent(query), {
                credentials: 'same-origin',
            }).then((x) => x.json()).then((result) => {
                if (instanceNum !== this.instanceNum) {
                    return;
                }
                if (result.error) {
                    handleError(result.error);
                    return;
                }
                result.forEach((store) => {
                    const elem = createListItem(store);
                    elem.addEventListener('click', () => {
                        this.close();
                        this.onAdd(store);
                    });
                    this.searchResults.appendChild(elem);
                });
            }).catch((err) => handleError(err));
        }
    }

    function createListItem(store) {
        const elem = document.createElement('li');
        elem.className = 'list-store';

        const logo = document.createElement('img');
        logo.className = 'logo';
        logo.src = 'svg/logo/' + store.source + '.svg';
        elem.appendChild(logo);

        const name = document.createElement('label');
        name.className = 'name';
        name.textContent = store.name;
        elem.appendChild(name);

        const address = document.createElement('label');
        address.className = 'address';
        address.textContent = store.address;
        elem.appendChild(address);

        return elem;
    }

    function handleError(err) {
        alert('Error: ' + err);
    }

    window.addEventListener('load', () => {
        window.storesPage = new StoresPage();
    });

})();