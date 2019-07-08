function updateList() {
    fetch('/list').then((x) => x.json()).then(updateListData);
}

function updateListData(data) {
    if (data.error) {
        alert(data.error);
        return;
    }
    const elem = document.getElementById('list-contents');
    elem.innerHTML = '';
    data.list.forEach((item, i) => {
        const zone = (data.zones[i] || {})['Name'] || '';
        const entry = document.createElement('div');
        entry.className = 'list-item';
        const name = document.createElement('label');
        name.className = 'list-name';
        name.textContent = item.name;
        const zoneName = document.createElement('label');
        zoneName.className = 'zone';
        zoneName.textContent = zone;
        entry.appendChild(name);
        entry.appendChild(zoneName);
        elem.appendChild(entry);
    });
}

function runSearch() {
    const query = document.getElementById('query').value;
    const url = '/search?query=' + encodeURIComponent(query);
    fetch(url).then((x) => x.json()).then(showSearchResults);
}

function showSearchResults(results) {
    if (results.error) {
        alert(results.error);
        return;
    }
    const elem = document.getElementById('search-results');
    elem.innerHTML = '';
    results.results.forEach((result, i) => {
        if (!result.inStock) {
            return;
        }
        const exactResult = results.exactResults[i];
        const photo = document.createElement('img');
        photo.className = 'photo';
        photo.src = result.photoUrl;
        const name = document.createElement('label');
        name.className = 'name';
        name.textContent = result.name;
        const add = document.createElement('button');
        add.addEventListener('click', () => {
            addListItem(exactResult);
        });
        add.textContent = 'Add';
        const entry = document.createElement('div');
        entry.className = 'search-result';
        entry.appendChild(photo);
        entry.appendChild(name);
        entry.appendChild(add);
        elem.appendChild(entry);
    });
}

function addListItem(rawData) {
    const postData = 'item=' + encodeURIComponent(JSON.stringify(rawData));
    fetch('/add', {
        method: 'POST',
        headers: {
            'content-type': 'application/x-www-form-urlencoded',
        },
        body: postData,
    }).then((x) => x.json()).then(updateListData);
}

window.addEventListener('load', () => {
    updateList();
    document.getElementById('search-button').addEventListener('click', runSearch);
});
