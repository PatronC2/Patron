import React from 'react';
import PayloadRow from './payloadrow'
const PayloadStruct = (props) => {

    const renderRows = () => (
        props.list ?
            props.list.map((row, i) => (
                <PayloadRow
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

export default PayloadStruct;