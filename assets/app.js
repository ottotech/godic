'use strict';

const e = React.createElement;

class DatabaseInfo extends React.Component {
    constructor(props) {
        super(props);
        this.state = { info: data["DatabaseInfo"] };
    }

    render() {

        return (
            <div>
                <p style={styles.p}><strong>Database name: </strong>{this.state.info["name"]}</p>
                <p style={styles.p}><strong>Database user: </strong>{this.state.info["user"]}</p>
                <p style={styles.p}><strong>Database host: </strong>{this.state.info["host"]}</p>
                <p style={styles.p}><strong>Database driver: </strong>{this.state.info["driver"]}</p>
                <p style={styles.p}><strong>Database schema: </strong>{this.state.info["schema"]}</p>
                <p style={styles.p}><strong>Database port: </strong>{this.state.info["port"]}</p>
            </div>
        );
    }
}

class TablesData extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            tables: [],
        };
        this.onChangeColumnDesc = this.onChangeColumnDesc.bind(this);
        this.onChangeTableDesc = this.onChangeTableDesc.bind(this);
    }

    componentDidMount() {
        let tables = data["Tables"];
        let columns = data["Columns"];

        for (let j = 0; j < tables.length; j++) {
            let cols = []

            for (let i = 0; i < columns.length; i++) {
                if (columns[i]["table_name"] === tables[j]["name"]) {
                    cols.push(columns[i])
                }
            }

            // we need to sort the cols accordingly, the rule is that
            // the PK will go in the first place and FKs will follow,
            // the rest of the columns will go afterwards.
            let pkIdx = false
            let fksIdx = []
            for (let w = 0; w < cols.length; w++) {
                if (cols[w]["is_primary_key"] === true) {
                    pkIdx = w
                } else if (cols[w]["is_foreign_key"] === true) {
                    fksIdx.push(w)
                }
            }

            if (typeof pkIdx === "number") {
                let pkObj = cols.splice(pkIdx, 1)[0]
                cols.splice(0, 0, pkObj)
            }

            for (let z = 0; z < fksIdx.length; z++) {
                let fkObj = cols.splice(fksIdx[z], 1)[0]
                cols.splice(1, 0, fkObj)
            }

            // finally we loop again through the table columns and check if any has a ENUM type.
            // If so, we need to create a custom description for that column.
            for (let k = 0; k < cols.length; k++) {
                if (cols[k]["has_enum"]) {
                    cols[k]["db_type"] = "ENUM("  + cols[k]["enum_values"].join() + ")"
                }
            }

            tables[j]["columns"] = cols
        }

        this.setState({tables:tables})
    }

    updateTableDictionary = (e) => {
        let tableIdx = e.target.getAttribute("data-table-idx");
        let tableName = e.target.getAttribute("data-table-name");
        let schema = window.location.protocol;
        let host = window.location.host;
        let endpoint = schema + "//" + host + "/update";
        let table = this.state.tables[tableIdx];
        let columns = table["columns"];
        let columnsData = [];

        // Let's ask the user if he/she really wants to update the data dictionary of the table.
        let yes = confirm("Are you sure you want to update the dictionary of table " + tableName +"?")
        if (!yes) {
            return
        }

        for (let i = 0; i < columns.length; i++) {
            let col = {}
            col["col_id"] = columns[i]["id"];
            col["description"] = columns[i]["description"];
            columnsData.push(col)
        }

        fetch(endpoint, {
            method: "POST",
            body: JSON.stringify({
                table_id: table["id"],
                table_description: table["description"],
                columns_data: columnsData
            })
        }).then(res => {
            if (res.status === 200) {
                alert("table " +  tableName + " has been updated successfully.")
            } else {
                res.text().then((text) => {
                    alert("An error occurred: " + text);
                })
            }
        }).catch(function (error) {
            console.log(error);
        });
    }

    onChangeTableDesc = (e) => {
        let tableIdx = e.target.getAttribute("data-table-idx");
        let tables = this.state.tables;
        tables[tableIdx]["description"] = e.target.value;
        this.setState({tables});
    }

    onChangeColumnDesc = (e) => {
        let tableIdx = e.target.getAttribute("data-table-idx");
        let colIdx = e.target.getAttribute("data-col-idx");
        let tables = this.state.tables;
        tables[tableIdx]["columns"][colIdx]["description"] = e.target.value;
        this.setState({tables});
    }

    rendeTables() {
        return this.state.tables.map((table, i) =>
            <Table
                key={i}
                tableIdx={i}
                tableName={table["name"]}
                tableID={table["id"]}
                tableDescription={table["description"]}
                tableColumns={table["columns"]}
                onChangeColumnDesc={this.onChangeColumnDesc}
                onChangeTableDesc={this.onChangeTableDesc}
                onClickSave={this.updateTableDictionary}
            />
        )
    }

    render() {
        return (
            <div>
                {this.rendeTables()}
                <TopBtn/>
            </div>
        );
    }
}

class Table extends React.Component{

    renderColumns = () => {
        return this.props.tableColumns.map((col, i) => {

            let key = ""
            if (col["is_primary_key"]) {
                key = "PK"
            } else if (col["is_foreign_key"]) {
                key = "FK"
            }

            let dbType = col["db_type"]
            if (col["db_type"].toUpperCase() === "VARCHAR") {
                dbType = dbType + "(" + col["length"] + ")"
             }


            let nullable = col["nullable"] === true ? "YES" : "NO"

            let unique = col["is_unique"] === true ? "YES" : "NO"

            return(
                <tr key={i}>
                    <td style={styles.table}>{key}</td>
                    <td style={styles.table}>{col["name"]}</td>
                    <td style={styles.table}>{dbType}</td>
                    <td style={styles.table}>{nullable}</td>
                    <td style={styles.table}>{unique}</td>
                    <td
                        data-table={col["table_name"]}
                        data-column-id={col["id"]}
                        style={styles.table}
                    >
                        <textarea
                            data-table-idx={this.props.tableIdx}
                            data-table={col["table_name"]}
                            data-col-id={col["id"]}
                            data-col-idx={i}
                            onChange={this.props.onChangeColumnDesc}
                            rows="5"
                            cols="50"
                            value={col["description"]}
                        />
                    </td>
                </tr>
            )
        })
    }

    render() {
        return (
            <div style={{marginTop: 50}}>
                <p style={styles.p}><strong>Table: </strong>{this.props.tableName}</p>
                <p style={styles.p}><strong>Description:</strong></p>
                <div style={{display: "flex"}}>
                    <textarea
                        data-table-idx={this.props.tableIdx}
                        onChange={this.props.onChangeTableDesc}
                        rows="4"
                        cols="80"
                        value={this.props.tableDescription}
                    />
                    <button
                        style={{width: 60}}
                        type="button"
                        data-table-name={this.props.tableName}
                        data-table-idx={this.props.tableIdx}
                        onClick={this.props.onClickSave}
                    >
                        save
                    </button>
                </div>
                <table style={styles.table}>
                    <thead>
                    <tr>
                        <th style={styles.table}>Key</th>
                        <th style={styles.table}>Attribute</th>
                        <th style={styles.table}>Data Type</th>
                        <th style={styles.table}>Nullable</th>
                        <th style={styles.table}>Unique</th>
                        <th style={styles.table}>Description</th>
                    </tr>
                    </thead>
                    <tbody>
                        {this.renderColumns()}
                    </tbody>
                </table>
            </div>
        );
    }
}


class TopBtn extends React.Component {
    constructor(props){
        super(props);
        this.state = {
            btn_display_style: "none",
            btn_opacity: 0.4
        };
        this.handleScroll = this.handleScroll.bind(this);
    }

    componentDidMount(){
        window.addEventListener('scroll', this.handleScroll);
    }

    componentWillUnmount() {
        window.removeEventListener('scroll', this.handleScroll);
    }

    handleScroll() {
        let bodyScrollPos = document.body.scrollTop;
        let docScrollPos = document.documentElement.scrollTop;
        if (bodyScrollPos > 20 || docScrollPos > 20){
            this.setState({ btn_display_style: "block" })
        }else {
            this.setState({ btn_display_style: "None" })
        }
    }

    render() {
        const styles = {};

        styles.top_btn = {
            display: this.state.btn_display_style,
            position: "fixed",
            bottom: 20,
            right: 30,
            zIndex: 99,
            fontSize: 18,
            border: "none",
            outline: "none",
            color: "grey",
            cursor: "pointer",
            padding: 15,
            borderRadius: 4,
            opacity: this.state.btn_opacity
        };

        return (
            <div>
                <button
                    style={styles.top_btn}
                    id="myBtn"
                    onClick={() => { document.documentElement.scrollTop = 0 }}
                    onMouseOver={() => this.setState({ btn_opacity: 0.8 }) }
                    onMouseOut={() => this.setState({ btn_opacity: 0.4 }) }
                >
                    Go to top
                </button>
            </div>
        )
    }
}


const styles = {
    p: {
        margin: 0
    },
    table: {
        border: "1px solid black"
    },
    saveBtn: {
        color: "white",
        padding: "15px 32px",
        textAlign: "center",
        fontSize: 16,
        cursor: "pointer",
        backgroundColor: "#008CBA",
        marginLeft: 3,
        outline: "none"
    }
}

const domContainer = document.querySelector('#databaseInfo');
const domContainerTwo = document.querySelector('#tablesData');
ReactDOM.render(e(DatabaseInfo), domContainer);
ReactDOM.render(e(TablesData), domContainerTwo);
