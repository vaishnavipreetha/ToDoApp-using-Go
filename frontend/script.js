const todoForm = document.getElementById('todo-form');
const todoList = document.getElementById('todo-list');
let editingTodoId = null;

// Fetch todos from the server
function fetchTodos() {
    fetch('/todos')
        .then(response => response.json())
        .then(todos => {
            todoList.innerHTML = '';
            todos.forEach(todo => {
                const li = document.createElement('li');
                li.innerHTML = `
                    <span>${todo.title} - ${todo.description}</span>
                    <button onclick="editTodo(${todo.id}, '${todo.title}', '${todo.description}')">Edit</button>
                    <button onclick="deleteTodo(${todo.id})">Delete</button>
                `;
                todoList.appendChild(li);
            });
        })
        .catch(error => console.error('Error fetching todos:', error));
}

// Add new todo
todoForm.addEventListener('submit', function(event) {
    event.preventDefault();
    const title = document.getElementById('title').value;
    const description = document.getElementById('description').value;

    if (editingTodoId) {
        // Update existing todo
        fetch(`/todos/${editingTodoId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ title, description, completed: false }),
        })
        .then(response => {
            if (response.ok) {
                fetchTodos(); // Refresh the todo list
                resetForm(); // Reset the form
            }
        })
        .catch(error => console.error('Error updating todo:', error));
    } else {
        // Create new todo
        fetch('/todos', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ title, description, completed: false }),
        })
        .then(response => {
            if (response.ok) {
                fetchTodos(); // Refresh the todo list
                resetForm(); // Reset the form
            }
        })
        .catch(error => console.error('Error adding todo:', error));
    }
});

// Edit todo
function editTodo(id, title, description) {
    editingTodoId = id;
    document.getElementById('title').value = title;
    document.getElementById('description').value = description;
}

// Delete todo
function deleteTodo(id) {
    fetch(`/todos/${id}`, {
        method: 'DELETE',
    })
    .then(response => {
        if (response.ok) {
            fetchTodos(); // Refresh the todo list
        }
    })
    .catch(error => console.error('Error deleting todo:', error));
}

// Reset form
function resetForm() {
    editingTodoId = null;
    todoForm.reset();
}

// Initial fetch of todos
fetchTodos();
