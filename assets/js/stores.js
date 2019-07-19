(function () {

    class StoresPage extends ListingPage {
        constructor() {
            super();
            this.updateData(window.STORES_DATA);
        }

        createAddDialog() {
            return new AddStoreDialog();
        }

        createListItem(store) {
            return createListItem(store);
        }

        selectedListItem(store) {
            window.open('/list?store=' + encodeURIComponent(store.id));
        }

        addItem(store) {
            const formData = 'source=' + encodeURIComponent(store.source) +
                '&signature=' + encodeURIComponent(store.signature) +
                '&data=' + encodeURIComponent(JSON.stringify(store.data));
            return fetch('/api/addstore', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {
                    'content-type': 'application/x-www-form-urlencoded',
                },
                body: formData,
            }).then((x) => x.json()).then((data) => {
                if (data.error) {
                    throw data.error;
                }
                store.id = data;
                this.addListItem(store);
            });
        }

        deleteItem(store) {
            return fetch('/api/removestore?store=' + encodeURIComponent(store.id), {
                credentials: 'same-origin',
            }).then((x) => x.json()).then((data) => {
                if (data.error) {
                    throw data.error;
                }
                this.updateData(data);
            });
        }
    }

    class AddStoreDialog extends AddDialog {
        fetchSearchResults(query) {
            return fetch('/api/storequery?query=' + encodeURIComponent(query), {
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
