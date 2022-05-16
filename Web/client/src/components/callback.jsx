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
        <header className="stuck">
            {renderRows()}
        </header>
    );
};

export default CommandBlock;