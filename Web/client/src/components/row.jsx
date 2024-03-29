import React, { Component } from 'react';
import { Link } from 'react-router-dom';

class Row extends Component {

    handleCheckboxChange = () => {
        const { onCheckboxChange, uuid } = this.props;
        onCheckboxChange(uuid);
      };

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
                        <input
                        type="checkbox"
                        checked={this.props.isSelected}
                        onChange={this.handleCheckboxChange}
                        />
                        <font size={2} color="#888888">
                            {" "}
                            {this.props.uuid.substring(0, 3)}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} style={{color: this.props.status == 'Online' ? '#00FF00' : '#FF3333'}}>
                            {/*Start News 1 Title */}{this.props.username}@{this.props.agentip}
                            {/*End News 1 Title */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Date */}
                            <Link to={`/configagent/${this.props.uuid}`}> configure </Link>
                            {/*End News 1 Date */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Date */}
                            <Link to={`/agent/${this.props.uuid}`}> interact </Link>
                            {/*End News 1 Date */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Date */}
                            <Link to={`/keylog/${this.props.uuid}`}>keylogs</Link>
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

export default Row;