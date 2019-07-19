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

        async addItem(store) {
            const formData = 'source=' + encodeURIComponent(store.source) +
                '&signature=' + encodeURIComponent(store.signature) +
                '&data=' + encodeURIComponent(JSON.stringify(store.data));
            const response = await fetch('/api/addstore', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {
                    'content-type': 'application/x-www-form-urlencoded',
                },
                body: formData,
            });
            const data = await response.json();
            if (data.error) {
                throw data.error;
            }
            store.id = data;
            this.addListItem(store);
        }

        async deleteItem(store) {
            const response = await fetch('/api/removestore?store=' + encodeURIComponent(store.id), {
                credentials: 'same-origin',
            });
            const data = await response.json();
            if (data.error) {
                throw data.error;
            }
            this.updateData(data);
        }
    }

    class AddStoreDialog extends AddDialog {
        async fetchSearchResults(query) {
            const response = await fetch('/api/storequery?query=' + encodeURIComponent(query), {
                credentials: 'same-origin',
            });
            const result = await response.json();
            if (result.error) {
                throw result.error;
            }
            return result;
        }

        createListItem(item) {
            return createListItem(item);
        }
    }

    function createListItem(store) {
        const elem = document.createElement('li');
        elem.className = 'list-item';

        const logo = document.createElement('img');
        logo.className = 'image';
        logo.src = 'svg/logo/' + store.source + '.svg';
        elem.appendChild(logo);

        const name = document.createElement('label');
        name.className = 'name';
        name.textContent = store.name;
        elem.appendChild(name);

        const address = document.createElement('label');
        address.className = 'location';
        address.textContent = store.address;
        elem.appendChild(address);

        return elem;
    }

    window.addEventListener('load', () => {
        window.storesPage = new StoresPage();
    });

})();
