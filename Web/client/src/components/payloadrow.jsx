import React, { Component } from 'react';
import { C2_ENDPOINT } from '../actions/c2actions'

class PayloadRow extends Component {



    renderRow = () => (
        <tr>
            <td>
                {/*Start News 1 Banner */}
                <pre>
                    <font face="lucida console">
                        <font size={2} color="#ff3333">
                            {" "}
                            ______________________________________________________{" "}
                        </font>
                        {"\n"}
                        <font size={2} color="#ff3333">
                            |
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}{this.props.concat}
                            {/*End News 1 Title */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +
                        </font>
                        {" "}
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}{this.props.serverip}:{this.props.serverport}
                            {/*End News 1 Title */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Date */}
                            <a href={`${C2_ENDPOINT}/files/${this.props.concat}`}>download </a>
                            {/*End News 1 Date */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +
                            <font size={2} color="#ff3333">
                                |
                            </font>
                            {"\n"}
                            <font size={2} color="#ff3333">
                                {" "}
                                ------------------------------------------------------{" "}
                            </font>
                        </font>
                    </font>
                </pre>
                {/*End News 1 Banner */}
            </td>
        </tr>
    )
    render() {

        return (
            this.renderRow()
        );
    }
}

export default PayloadRow;