import React, { Component } from 'react';
import { Link } from 'react-router-dom';

class AgentByIpRow extends Component {


    renderRow = () => (
        <tr>
            <td>
                {/*Start News 1 Banner */}
                <pre>
                    <font face="lucida console">
                        <font size={2} color="#ff3333">
                            {" "}
                            ____________________________________________________________{" "}
                        </font>
                        {"\n"}
                        <font size={2} color="#ff3333">
                            |
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} style={{color: '#00FF00'}}>
                            {/*Start News 1 Title */}
                            {/*End News 1 Title */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Date */}
                            <Link to={`/groupagent/${this.props.agentip}`}> {this.props.agentip} </Link>
                            {/*End News 1 Date */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Date */}
                            {" "}
                            {/*End News 1 Date */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Date */}
                            
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
                                ------------------------------------------------------------{" "}
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

export default AgentByIpRow;