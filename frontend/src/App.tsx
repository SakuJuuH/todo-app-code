import {useEffect, useState} from 'react'
import axios from 'axios'
import './App.css'

interface ImageInfo {
    path: string;
    cached_at: string;
}

interface Todo {
    id: number;
    task: string
    done: boolean
}

function App() {
    const [imageInfo, setImageInfo] = useState<ImageInfo | null>(null);
    const [imageLoading, setImageLoading] = useState(true);
    const [imageError, setImageError] = useState<string | null>(null);
    const [todos, setTodos] = useState<Todo[]>([]);
    const [todosLoading, setTodosLoading] = useState(false);
    const [todosError, setTodosError] = useState<string | null>(null);

    const [todoTask, setTodoTask] = useState<string>('');

    const imageServiceUrl = import.meta.env.PROD
        ? '/api/image'
        : 'http://localhost:3000/api/image'; // for local development

    const todoServiceUrl = import.meta.env.PROD
        ? '/api/todos'
        : 'http://localhost:3001/api/todos'; // for local development

    useEffect(() => {
        const fetchImageInfo = async () => {
            try {
                setImageLoading(true)
                setImageError(null);

                const response = await axios.get<ImageInfo>(`${imageServiceUrl}/current`);
                const data: ImageInfo = response.data;
                if (!data || !data.path) {
                    throw new Error('Invalid image data received');
                }

                let image: ImageInfo = {
                    path: data.path,
                    cached_at: data.cached_at,
                }

                setImageInfo(image);

            } catch (error) {
                console.error('Error fetching image:', error);
                setImageError('Failed to fetch image');
            } finally {
                setImageLoading(false);
            }
        };

        const fetchTodoItems = async () => {
            try {
                setTodosLoading(true);
                setTodosError(null);

                const response = await axios.get<Todo[]>(`${todoServiceUrl}`);
                const data: Todo[] = response.data;

                if (!data || !Array.isArray(data)) {
                    throw new Error('Invalid todos data received');
                }

                const todos: Todo[] = data.map((item: Todo) => ({
                    id: item.id,
                    task: item.task,
                    done: item.done,
                }));

                setTodos(todos);
            } catch (error) {
                console.error('Error fetching todos:', error);
                setTodosError('Failed to fetch todos');
            } finally {
                setTodosLoading(false);
            }
        };

        fetchImageInfo().then();
        fetchTodoItems().then();
    }, [imageServiceUrl, todoServiceUrl]);

    const handleShutdown = async () => {
        try {
            await axios.post(`${imageServiceUrl}/shutdown`);
            alert('Server shutdown initiated!');
        } catch (error) {
            console.error('Error shutting down server:', error);
        }
    };

    const handleAddTodo = async () => {
        if (!todoTask.trim()) {
            alert('Please enter a todo item.');
            return;
        }

        if (todoTask.length > 140) {
            alert('Todo item cannot exceed 140 characters.');
            return;
        }

        try {
            console.log(`Adding todo: ${todoTask}`);

            const response = await axios.post<Todo>(`${todoServiceUrl}`, {task: todoTask});
            const data: Todo = response.data

            let newTodo: Todo = {
                id: data.id,
                task: data.task,
                done: data.done,
            }

            console.log(`Todo added:  ${newTodo.id} - ${newTodo.task} (${newTodo.done})`);

            let prevTodos = todos || [];

            setTodos([...prevTodos, newTodo]);
            setTodoTask('');
        } catch (error) {
            console.error('Error adding todo:', error);
            alert('Failed to add todo item.');
        }
    };

    const handleCompleteTodo = async (id: number) => {
        try {
            await axios.put(`${todoServiceUrl}/${id}`);

            let prevTodos = todos || [];
            let updatedTodos = prevTodos.map(todo =>
                todo.id === id ? {...todo, done: true} : todo
            );
            setTodos(updatedTodos);
        } catch (error) {
            console.error('Error completing todo:', error);
            alert('Failed to complete todo item.');
        }
    };

    return (
        <>
            <div>
                <h1>The Todo App</h1>
                {imageLoading && <p>Loading image...</p>}
                {imageError && <p style={{color: 'red'}}>{imageError}</p>}
                {imageInfo && !imageLoading && (
                    <div>
                        <p>Image cached at: {new Date(imageInfo.cached_at).toLocaleString()}</p>
                        <img
                            src={`${imageServiceUrl}${imageInfo.path}`}
                            alt="Random image from Picsum"
                            style={{maxWidth: '100%', height: 'auto'}}
                        />
                    </div>
                )}
                <div className="todo">
                    <input
                        className="input"
                        type="text"
                        placeholder=""
                        value={todoTask}
                        onSubmit={handleAddTodo}
                        onChange={(e) => setTodoTask(e.target.value)}
                    />
                    <button className="button" onClick={handleAddTodo}>
                        Add Todo
                    </button>
                </div>
                {todosLoading && <p>Loading todos...</p>}
                {todosError && <p style={{color: 'red'}}>{todosError}</p>}
                {todos && !todosLoading && (
                    <div className="todo-lists-container">
                        <div className="todo-list uncompleted-list">
                            <h3>Uncompleted Todos</h3>
                            {todos.sort((a, b) => a.id - b.id).filter(todo => !todo.done).map((todo) => (
                                <div key={todo.id} className="todo-item">
                                    {todo.id}. {todo.task}
                                    <button className='todo-button' onClick={() => handleCompleteTodo(todo.id)}>
                                        ✓
                                    </button>
                                </div>
                            ))}
                            {todos.every(todo => todo.done) && <p>No pending todos. Add a new todo!</p>}
                        </div>
                        <div className="todo-list completed-list">
                            <h3>Completed Todos</h3>
                            {todos.sort((a, b) => a.id - b.id).filter(todo => todo.done).map((todo) => (
                                <div key={todo.id} className="todo-item completed">
                                    {todo.id}. {todo.task}
                                </div>
                            ))}
                        </div>
                    </div>
                )}
                <div style={{marginTop: '20px'}}>
                    <button onClick={handleShutdown}>
                        Shutdown Server (for testing)
                    </button>
                </div>
            </div>
        </>
    );
}

export default App;