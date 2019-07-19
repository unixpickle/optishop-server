(function () {

    class StoresPage {
        constructor() {
            this.addButton = document.getElementById('add-button');
            this.storesList = document.getElementById('stores-list');
            this.noStores = document.getElementById('no-stores');
            this.addStoreDialog = new AddStoreDialog();

            this.addButton.addEventListener('click', () => this.addStoreDialog.open());
            this.addStoreDialog.onAdd = (store) => this.addStore(store);

            this.data = null;
            this.updateData(window.STORES_DATA);
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

    class AddStoreDialog extends AddDialog {
        fetchSearchResults(query) {
            fetch('/api/storequery?query=' + encodeURIComponent(query), {
                credentials: 'same-origin',
            }).then((x) => x.json()).then((result) => {
                if (result.error) {
                    throw result.error;
                }
                return result;
            }).catch((err) => handleError(err));
        }

        createListItem(item) {
            return createListItem(item);
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

    window.addEventListener('load', () => {
        window.storesPage = new StoresPage();
    });

})();
