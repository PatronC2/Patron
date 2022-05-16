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
        <header className="stuck">
            {renderRows()}
        </header>
    );
};

export default PayloadStruct;