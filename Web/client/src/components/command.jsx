import React from 'react';
import CommandRow from './commandrow'
const CommandBlock = (props) => {

    const renderRows = () => (
        props.list ?
            props.list.map((row, i) => (
                <CommandRow
                    {...row}
                />
            ))
            : null
    )

    return (
        <React.Fragment>
            {renderRows()}
        </React.Fragment>
    );
};

export default CommandBlock;