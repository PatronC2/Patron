import React from 'react';
import Row from './row'
const CommandBlock = (props) => {

    const renderRows = () => (
        props.list ?
            props.list.map((row, i) => (
                <Row
                    {...row}
                />
            ))
            : null
    )

    return (
        <header className="App-header">
            {renderRows()}
        </header>
    );
};

export default CommandBlock;