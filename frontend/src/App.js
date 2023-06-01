import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Table, Tbody, Td, Th, Thead, Tr } from 'radix-ui';

function App() {
  const [results, setResults] = useState([]);

  useEffect(() => {
    const interval = setInterval(() => {
      axios.get('http://localhost:8080/')
          .then(response => {
            setResults(response.data);
          });
    }, 5000);
    return () => clearInterval(interval);
  }, []);

  return (
      <div className="App">
        <Table>
          <Thead>
            <Tr>
              <Th>Website</Th>
              <Th>Status</Th>
              <Th>Timestamp</Th>
            </Tr>
          </Thead>
          <Tbody>
            {results.map(result => (
                <Tr key={result.id}>
                  <Td>{result.website}</Td>
                  <Td>{result.status}</Td>
                  <Td>{result.timestamp}</Td>
                </Tr>
            ))}
          </Tbody>
        </Table>
      </div>
  );
}

export default App;