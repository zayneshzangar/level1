let currentParentId = null;
let searchQuery = '';

// Показ ошибок
function showError(message) {
    document.getElementById('error').innerText = message;
}

function clearError() {
    document.getElementById('error').innerText = '';
}

// Загрузка комментариев
function loadComments(page = 1, parentId = null) {
    clearError();
    currentParentId = parentId;
    const url = parentId 
        ? `/comments?parent=${parentId}` 
        : `/comments?page=${page}&limit=10&sort=asc${searchQuery ? '&search=' + encodeURIComponent(searchQuery) : ''}`;
    console.log('Запрос к API:', url);
    fetch(url)
        .then(res => {
            if (!res.ok) {
                throw new Error(`HTTP ${res.status}: ${res.statusText}`);
            }
            return res.json();
        })
        .then(data => {
            console.log('Данные от API:', data);
            document.getElementById('commentsTree').innerHTML = renderTree(data.comments || [], 0);
            if (data.pages > 1 && !parentId) {
                let nav = '<div>Страницы: ';
                for (let p = 1; p <= data.pages; p++) {
                    nav += `<button onclick="loadComments(${p})">${p}</button> `;
                }
                nav += '</div>';
                document.getElementById('commentsTree').innerHTML += nav;
            }
        })
        .catch(err => {
            showError('Ошибка загрузки: ' + err.message);
            console.error('Ошибка:', err);
        });
}

// Рендер дерева
function renderTree(comments, level = 0) {
    if (!comments || comments.length === 0) return '<p>Нет комментариев.</p>';
    let html = '';
    comments.forEach(comment => {
        const indent = '&nbsp;'.repeat(level * 3) + (level > 0 ? '↳ ' : '');
        const created = new Date(comment.created_at).toLocaleString('ru');
        html += `
            <div class="comment" style="margin-left: ${level * 20}px;">
                <div class="comment-header">
                    ${indent}ID: ${comment.id} | ${created}
                    <span class="delete-btn" onclick="deleteComment(${comment.id})">[Удалить]</span>
                    <span onclick="setReplyTo(${comment.id})" style="cursor: pointer; color: blue;">[Ответить]</span>
                </div>
                <div class="comment-content">${comment.content}</div>
                ${renderTree(comment.children || [], level + 1)}
            </div>
        `;
    });
    return html;
}

// Создание комментария
function createComment() {
    clearError();
    const content = document.getElementById('newContent').value.trim();
    if (!content) {
        showError('Введите текст!');
        return;
    }
    const parentId = document.getElementById('replyToId').value || null;
    const payload = { content };
    if (parentId) payload.parent_id = parseInt(parentId);
    
    console.log('Создание комментария:', payload);
    fetch('/comments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    })
    .then(res => {
        if (!res.ok) {
            throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }
        return res.json();
    })
    .then(data => {
        console.log('Создан:', data);
        document.getElementById('newContent').value = '';
        document.getElementById('replyToId').value = '';
        loadComments(1, currentParentId);
    })
    .catch(err => {
        showError('Ошибка создания: ' + err.message);
        console.error('Ошибка:', err);
    });
}

// Удаление
function deleteComment(id) {
    if (!confirm('Удалить и все ответы?')) return;
    fetch(`/comments/${id}`, { method: 'DELETE' })
        .then(res => {
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            return res.json();
        })
        .then(() => loadComments(1, currentParentId))
        .catch(err => {
            showError('Ошибка удаления: ' + err.message);
        });
}

// Поиск
function searchComments() {
    searchQuery = document.getElementById('searchInput').value.trim();
    loadComments(1);
}

// Установка для ответа
function setReplyTo(id) {
    document.getElementById('replyToId').value = id;
    alert('Выбран для ответа ID: ' + id);
}

function replyToSelected() {
    const parentId = document.getElementById('replyToId').value;
    if (!parentId) return alert('Сначала нажмите [Ответить]!');
    alert('Ответ будет на ID: ' + parentId);
}

// Инициализация
loadComments();