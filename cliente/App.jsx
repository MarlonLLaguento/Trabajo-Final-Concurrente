import React, { useState, useEffect } from 'react';
import './App.css'

function App() {
    const [inputValue, setInputValue] = useState('');
    const [recommendations, setRecommendations] = useState([]);
    const [mode, setMode] = useState(null);
    const [socket, setSocket] = useState(null);
    const [hasPosters, setHasPosters] = useState(false); 
    const omdbApiKey = 'bb7c100b'; 

    useEffect(() => {
        const createWebSocket = () => {
            const ws = new WebSocket('ws://127.0.0.1:10001/ws');

            ws.onopen = () => {
                console.log('Conexión WebSocket establecida');
                setSocket(ws);
            };

            ws.onmessage = (event) => {
                const data = JSON.parse(event.data);
                setRecommendations(data);
                setHasPosters(false); 
            };

            ws.onerror = (error) => {
                console.error('Error en la conexión WebSocket:', error);
            };

            ws.onclose = () => {
                console.log('Conexión WebSocket cerrada, reintentando en 3 segundos...');
                setTimeout(createWebSocket, 3000); // Reintentar cada 3 segundos si se cierra la conexión
            };
        };

        createWebSocket();

        return () => {
            if (socket) {
                socket.close();
            }
        };
    }, []);

    const handleInputChange = (e) => {
        setInputValue(e.target.value);
    };

    const handleModeChange = (selectedMode) => {
        setMode(selectedMode);
        setInputValue('');
        setRecommendations([]);
        setHasPosters(false); 
    };

    const handleSearch = () => {
        if (socket && socket.readyState === WebSocket.OPEN) {
            let message;
            if (mode === 'title') {
                message = JSON.stringify({ send: 1, opc: 1, movGenre: [inputValue] });
            } else if (mode === 'genres') {
                const genres = inputValue.split(',').map(genre => genre.trim());
                message = JSON.stringify({ send: 2, opc: 2, movGenre: genres });
            }

            if (message) {
                socket.send(message);
            }
        }
    };

    const formatTitleForUrl = (title) => {
        let formattedTitle = title
            .replace(/, The/g, '')             
            .replace(/\s+/g, '+')     
            .replace(/\(\d{4}\)/, '')
            .replace(/^(\+)+|(\+)+$/g, ''); 
            // .replace(/[^a-zA-Z0-9()+\- ]/g, '');
    
        if (title.includes(', The')) {
            formattedTitle = 'The+' + formattedTitle;
        }
        //console.log(formattedTitle);
        return formattedTitle.trim();
    };

    const getMoviePoster = async (title) => {
        try {
            const formattedTitle = formatTitleForUrl(title);
            const response = await fetch(`https://www.omdbapi.com/?s=${formattedTitle}&apikey=${omdbApiKey}`);
            const data = await response.json();
            
            if (data.Search && data.Search.length > 0) {
                // Verificar si la película tiene póster
                if (data.Search[0].Poster !== 'N/A') {
                    return data.Search[0].Poster;
                } else {
                    return 'https://via.placeholder.com/100x150?text=No+Image';
                }
            } else {
                return 'https://via.placeholder.com/100x150?text=No+Image'; // Si no se encuentra la película
            }
        } catch (error) {
            console.error('Error al obtener el poster de OMDB:', error);
            return 'https://via.placeholder.com/100x150?text=No+Image';
        }
    };    

    useEffect(() => {
        if (recommendations.length > 0 && !hasPosters) {
            const addPostersToRecommendations = async () => {
                const updatedRecommendations = await Promise.all(recommendations.map(async (rec) => {
                    const poster = await getMoviePoster(rec.Title);
                    return { ...rec, Poster: poster };
                }));
                setRecommendations(updatedRecommendations);
                setHasPosters(true);
            };

            addPostersToRecommendations();
        }
    }, [recommendations, hasPosters]);

    return (
        <div>
            <h1>ReMovie</h1>
            <h2>Recomendaciones de Películas</h2>
            <div>
                <button onClick={() => handleModeChange('title')}>Buscar por Título</button>
                <button onClick={() => handleModeChange('genres')}>Buscar por Géneros</button>
            </div>

            {mode && (
                <div>
                    <h2>{mode === 'title' ? 'Ingrese el Título de la Película' : 'Ingrese los Géneros de la Película'}</h2>
                    <input
                        type="text"
                        value={inputValue}
                        onChange={handleInputChange}
                        placeholder={mode === 'title' ? 'Ej: Toy Story' : 'Ej: action, crime'}
                    />
                    <button onClick={handleSearch}>Buscar</button>
                </div>
            )}

            {recommendations.length > 0 && (
                /*<div>
                    <h2>Recomendaciones:</h2>
                    <ol>
                        {recommendations.map((rec, index) => (
                            <li key={index}>
                                <strong>Título:</strong> {rec.Title}, <strong>Géneros:</strong> {rec.Genres.join(', ')},
                                <strong> Calificación Promedio:</strong> {rec.AvgRating.toFixed(2)},
                                <strong> Número de Calificaciones:</strong> {rec.NumRatings}
                            </li>
                        ))}
                    </ol>
                </div> */
                <div>
                    <h2>Recomendaciones:</h2>
                    <table style={{ width: '100%', borderCollapse: 'collapse', marginTop: '10px' }}>
                        <thead>
                            <tr>
                                <th style={{ border: '1px solid #ddd', padding: '8px', backgroundColor: '#f2f2f2' }}>Póster</th>
                                <th style={{ border: '1px solid #ddd', padding: '8px', backgroundColor: '#f2f2f2' }}>Título</th>
                                <th style={{ border: '1px solid #ddd', padding: '8px', backgroundColor: '#f2f2f2' }}>Géneros</th>
                                <th style={{ border: '1px solid #ddd', padding: '8px', backgroundColor: '#f2f2f2' }}>Calificación Promedio</th>
                                <th style={{ border: '1px solid #ddd', padding: '8px', backgroundColor: '#f2f2f2' }}>Número de Calificaciones</th>
                            </tr>
                        </thead>
                        <tbody>
                            {recommendations.map((rec, index) => (
                                <tr key={index} style={{ borderBottom: '1px solid #ddd' }}>
                                    <td style={{ border: '1px solid #ddd', padding: '8px' }}>
                                        <img src={rec.Poster} alt={rec.Title} style={{ width: '100px', height: '150px', objectFit: 'cover' }} />
                                    </td>
                                    <td style={{ border: '1px solid #ddd', padding: '8px' }}>{rec.Title}</td>
                                    <td style={{ border: '1px solid #ddd', padding: '8px' }}>{rec.Genres.join(', ')}</td>
                                    <td style={{ border: '1px solid #ddd', padding: '8px' }}>{rec.AvgRating.toFixed(2)}</td>
                                    <td style={{ border: '1px solid #ddd', padding: '8px' }}>{rec.NumRatings}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

            )}
        </div>
    );
}

export default App;
