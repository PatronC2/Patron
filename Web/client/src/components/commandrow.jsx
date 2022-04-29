import React, { Component } from 'react';

class CommandRow extends Component {

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
                        {/* <font size={2} color="#ff3333">
                            |
                        </font> */}
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <textarea value={this.props.command} rows={1}/>
                            {/*End News 1 Title */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{"     "}
                        </font>
                        <font size={2} color="#888888">
                            {/* <font size={2} color="#ff3333">
                                |
                            </font> */}
                            {"\n"}
                            <font size={2} color="#ff3333">
                                {" "}
                                ------------------------------------------------------{" "}
                            </font>
                        </font>
                    </font>
                </pre>
                {/*End News 1 Banner */}
                {/*Start News 1 Banner */}
                <pre>
                    <font face="lucida console">
                        <font size={2} color="#00FFFF">
                            {" "}
                            ______________________________________________________{" "}
                        </font>
                        {"\n"}
                        {/* <font size={2} color="#ff3333">
                            |
                        </font> */}
                        <font size={2} color="#888888">
                            {" "}
                            +{" "}
                        </font>
                        <font size={2} color="#FFFFFF">
                            {/*Start News 1 Title */}
                            {/* {this.props.command} */}
                            <textarea value={this.props.output ? this.props.output : "Loading..." } rows={5}/>
                            {/*End News 1 Title */}
                        </font>
                        <font size={2} color="#888888">
                            {" "}
                            +{"     "}
                        </font>
                        <font size={2} color="#888888">
                            {/* <font size={2} color="#ff3333">
                                |
                            </font> */}
                            {"\n"}
                            <font size={2} color="#00FFFF">
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

export default CommandRow;