'use strict';

const e = React.createElement;

class DatabaseInfo extends React.Component {
    constructor(props) {
        super(props);
        console.log(data)
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
    }

    componentDidMount() {
        let tables = data["Tables"]
        let columns = data["Columns"]

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

            tables[j]["columns"] = cols
        }

        this.setState({tables: tables})
    }

    onChangeColumnDesc = (e) => {
        let tableName = e.target.getAttribute("data-table");
        let colID = e.target.getAttribute("data-col-id");
        let colIdx = e.target.getAttribute("data-col-idx");
    }

    rendeTables() {
        return this.state.tables.map((table, i) =>
            <Table
                key={i}
                tableName={table["name"]}
                tableID={table["id"]}
                tableColumns={table["columns"]}
                onChangeColumnDesc={this.onChangeColumnDesc}
            />
        )
    }

    render() {
        return (
            <div>
                {this.rendeTables()}
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
                            data-table={col["table_name"]}
                            data-col-id={col["id"]}
                            data-col-idx={i}
                            onChange={this.props.onChangeColumnDesc}
                            rows="4"
                            cols="50"
                        />
                    </td>
                    <td>
                        <button type="button">save</button>
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
                    <textarea id={this.props.tableID} rows="4" cols="80"/>
                    <button type="button">save</button>
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
                        <th style={styles.table}>Action</th>
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
